package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/common/gin/response"
	"kcaitech.com/kcserver/common/jwt"
	"kcaitech.com/kcserver/common/models"
	"kcaitech.com/kcserver/common/services"
	config "kcaitech.com/kcserver/controllers"

	// "kcaitech.com/kcserver/utils/sliceutil"
	"strings"
	"time"

	"kcaitech.com/kcserver/utils/str"
	. "kcaitech.com/kcserver/utils/time"
)

type wxOpenWebLoginReq struct {
	Id         string `json:"id"`
	Code       string `json:"code" binding:"required"`
	InviteCode string `json:"invite_code"`
}

var InviteCodeList = []string{ // 邀请码
	"fo3yblC5",
	"2gampt0q",
}

const IsInviteCodeCheck = false // 是否开启邀请码校验

type wxOpenWebLoginResp struct {
	Id       string `json:"id"`
	Nickname string `json:"nickname"`
	Token    string `json:"token"`
	Avatar   string `json:"avatar"`
}

func (resp *wxOpenWebLoginResp) MarshalJSON() ([]byte, error) {
	if strings.HasPrefix(resp.Avatar, "/") {
		resp.Avatar = config.Config.StorageUrl.Attatch + resp.Avatar
	}
	return json.Marshal(struct {
		wxOpenWebLoginResp
	}{wxOpenWebLoginResp: *resp})
}

type wxOpenWebAccessTokenResp struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Openid       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionId      string `json:"unionid"`
}

type wxOpenWebUserInfoResp struct {
	Openid     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	Headimgurl string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	UnionId    string   `json:"unionid"`
}

// WxOpenWebLogin 微信开放平台网站登录
func WxOpenWebLogin(c *gin.Context) {
	var req wxOpenWebLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userService := services.NewUserService()
	user := &models.User{}

	if req.Id != "" {
		if err := userService.Get(user, "id = ? and wx_login_code = ?", req.Id, req.Code); err != nil {
			response.BadRequest(c, "参数错误：id")
			return
		}
	} else {
		// Code换取AccessToken
		// 发起请求
		queryParams := url.Values{}
		queryParams.Set("appid", config.Config.Wx.Appid)
		queryParams.Set("secret", config.Config.Wx.Secret)
		queryParams.Set("code", req.Code)
		queryParams.Set("grant_type", "authorization_code")
		resp, err := http.Get("https://api.weixin.qq.com/sns/oauth2/access_token" + "?" + queryParams.Encode())
		if err != nil {
			log.Println(err)
			response.Fail(c, "登陆失败")
			return
		}
		defer resp.Body.Close()
		// 读取响应
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			response.Fail(c, "登陆失败")
			return
		}
		// json解析
		var wxAccessTokenResp wxOpenWebAccessTokenResp
		err = json.Unmarshal(body, &wxAccessTokenResp)
		if err != nil {
			log.Println(err)
			response.Fail(c, "登陆失败")
			return
		}
		if wxAccessTokenResp.Openid == "" || wxAccessTokenResp.UnionId == "" {
			log.Println("OpenId或UnionId为空", string(body))
			response.Fail(c, "登陆失败")
			return
		}

		// AccessToken换取用户信息
		// 发起请求
		queryParams = url.Values{}
		queryParams.Set("access_token", wxAccessTokenResp.AccessToken)
		queryParams.Set("openid", wxAccessTokenResp.Openid)
		queryParams.Set("lang", "lang")
		resp, err = http.Get("https://api.weixin.qq.com/sns/userinfo" + "?" + queryParams.Encode())
		if err != nil {
			log.Println(err)
			response.Fail(c, "登陆失败")
			return
		}
		defer resp.Body.Close()
		// 读取响应
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			response.Fail(c, "登陆失败")
			return
		}
		// json解析
		var wxUserInfoResp wxOpenWebUserInfoResp
		err = json.Unmarshal(body, &wxUserInfoResp)
		if err != nil {
			log.Println(err)
			response.Fail(c, "登陆失败")
			return
		}
		if wxUserInfoResp.Nickname == "" {
			log.Println("Nickname或Headimgurl为空")
			response.Fail(c, "登陆失败")
			return
		}

		err = userService.Get(user, "wx_open_id = ?", wxAccessTokenResp.Openid)
		if err != nil && !errors.Is(err, services.ErrRecordNotFound) {
			log.Println("userService.Get错误：", err)
			response.Fail(c, "登陆失败")
			return
		}
		if errors.Is(err, services.ErrRecordNotFound) {
			// 创建用户
			t := Time(time.Now())
			user = &models.User{
				Nickname:                 wxUserInfoResp.Nickname,
				WxOpenId:                 wxAccessTokenResp.Openid,
				WxUnionId:                wxAccessTokenResp.UnionId,
				WxAccessToken:            wxAccessTokenResp.AccessToken,
				WxAccessTokenCreateTime:  t,
				WxRefreshToken:           wxAccessTokenResp.RefreshToken,
				WxRefreshTokenCreateTime: t,
				Avatar:                   wxUserInfoResp.Headimgurl,
				Uid:                      str.GetUid(),
				WxLoginCode:              req.Code,
				IsActivated:              !IsInviteCodeCheck,
			}
			err := userService.Create(user)
			if err != nil {
				log.Println(err)
				response.Fail(c, "登陆失败")
				return
			}
			// 下载微信头像
			go func(user *models.User, avatarUrl string) {
				resp, err := http.Get(avatarUrl)
				if err != nil {
					log.Println("下载头像失败：", err)
					return
				}
				defer resp.Body.Close()
				fileBytes := make([]byte, resp.ContentLength)
				if _, err := resp.Body.Read(fileBytes); err != nil {
					response.BadRequest(c, "读取文件失败")
					return
				}
				if _, err := services.NewUserService().UploadUserAvatar(user, fileBytes, resp.Header.Get("Content-Type")); err != nil {
					log.Println("上传头像失败：", err)
					return
				}
			}(user, wxUserInfoResp.Headimgurl)
		} else {
			user.WxUnionId = wxAccessTokenResp.UnionId
			if _, err := userService.UpdatesIgnoreZero(user); err != nil {
				log.Println("更新UnionId失败：", err)
			}
		}
	}

	// if IsInviteCodeCheck && !user.IsActivated {
	// 	user.WxLoginCode = req.Code
	// 	_, _ = userService.Updates(user)
	// 	// 邀请码校验
	// 	if len(sliceutil.FilterT(func(code string) bool {
	// 		return req.InviteCode == code
	// 	}, InviteCodeList...)) == 0 {
	// 		//response.BadRequest(c, "邀请码错误")
	// 		response.Resp(c, http.StatusBadRequest, "邀请码错误", map[string]interface{}{"id": str.IntToString(user.Id)})
	// 		return
	// 	}
	// 	if err := services.NewInviteCodeService().Create(&models.InviteCode{
	// 		Code:   req.InviteCode,
	// 		UserId: user.Id,
	// 	}); err != nil {
	// 		log.Println("创建邀请码记录失败：", err)
	// 		//response.BadRequest(c, "邀请码错误")
	// 		response.Resp(c, http.StatusBadRequest, "邀请码错误", map[string]interface{}{"id": str.IntToString(user.Id)})
	// 		return
	// 	}
	// 	// 激活用户
	// 	user.IsActivated = true
	// 	if _, err := userService.Updates(user); err != nil {
	// 		response.Fail(c, "登陆失败")
	// 		return
	// 	}
	// }

	// 创建JWT
	token, err := jwt.CreateJwt(&jwt.Data{
		Id:       str.IntToString(user.Id),
		Nickname: user.Nickname,
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, &wxOpenWebLoginResp{
		Id:       str.IntToString(user.Id),
		Nickname: user.Nickname,
		Token:    token,
		Avatar:   user.Avatar,
	})
}
