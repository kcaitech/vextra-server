package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"net/url"
	"protodesign.cn/kcserver/authservice/config"
	"protodesign.cn/kcserver/common"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/jwt"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/utils/sliceutil"
	"protodesign.cn/kcserver/utils/str"
	. "protodesign.cn/kcserver/utils/time"
	"strings"
	"time"
)

type wxLoginReq struct {
	Id         string `json:"id"`
	Code       string `json:"code" binding:"required"`
	InviteCode string `json:"invite_code"`
}

var InviteCodeList = []string{
	"fo3yblC5",
	"2gampt0q",
	"d2z8ARv6",
	"5oO63m7R",
	"eDKn3m9P",
	"69B833mt",
	"rT4bKQ3h",
	"QlIYCRWf",
	"3pvn803r",
	"84L0w5dS",
}

type wxLoginResp struct {
	Id       string `json:"id"`
	Nickname string `json:"nickname"`
	Token    string `json:"token"`
	Avatar   string `json:"avatar"`
}

func (resp *wxLoginResp) MarshalJSON() ([]byte, error) {
	if strings.HasPrefix(resp.Avatar, "/") {
		resp.Avatar = common.StorageHost + resp.Avatar
	}
	return json.Marshal(struct {
		wxLoginResp
	}{wxLoginResp: *resp})
}

type wxAccessTokenResp struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Openid       string `json:"openid"`
	Scope        string `json:"scope"`
	Unionid      string `json:"unionid"`
}

type wxUserInfoResp struct {
	Openid     string   `json:"openid"`
	Nickname   string   `json:"nickname"`
	Sex        int      `json:"sex"`
	Province   string   `json:"province"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	Headimgurl string   `json:"headimgurl"`
	Privilege  []string `json:"privilege"`
	Unionid    string   `json:"unionid"`
}

// WxLogin 微信登录
func WxLogin(c *gin.Context) {
	var req wxLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	userService := services.NewUserService()

	if req.Id != "" {
		user := &models.User{}
		if err := userService.Get(user, "id = ? and wx_login_code = ?", req.Id, req.Code); err != nil {
			response.BadRequest(c, "参数错误：id")
			return
		}
		// 创建JWT
		token, err := jwt.CreateJwt(&jwt.Data{
			Id:       str.IntToString(user.Id),
			Nickname: user.Nickname,
		})
		if err != nil {
			response.Fail(c, err.Error())
			return
		}
		response.Success(c, &wxLoginResp{
			Id:       str.IntToString(user.Id),
			Nickname: user.Nickname,
			Token:    token,
			Avatar:   user.Avatar,
		})
		return
	}

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
	var wxAccessTokenResp wxAccessTokenResp
	err = json.Unmarshal(body, &wxAccessTokenResp)
	if err != nil {
		log.Println(err)
		response.Fail(c, "登陆失败")
		return
	}
	if wxAccessTokenResp.Openid == "" {
		log.Println("OpenId为空")
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
	var wxUserInfoResp wxUserInfoResp
	err = json.Unmarshal(body, &wxUserInfoResp)
	if err != nil {
		log.Println(err)
		response.Fail(c, "登陆失败")
		return
	}
	if wxUserInfoResp.Nickname == "" || wxUserInfoResp.Headimgurl == "" {
		log.Println("Nickname或Headimgurl为空")
		response.Fail(c, "登陆失败")
		return
	}

	user := &models.User{}
	err = userService.Get(user, "wx_open_id = ?", wxAccessTokenResp.Openid)
	if err != nil && err != services.ErrRecordNotFound {
		log.Println("userService.Get错误：", err)
		response.Fail(c, "登陆失败")
		return
	}
	if err == services.ErrRecordNotFound {
		// 创建用户
		t := Time(time.Now())
		user = &models.User{
			Nickname:                 wxUserInfoResp.Nickname,
			WxOpenId:                 wxAccessTokenResp.Openid,
			WxAccessToken:            wxAccessTokenResp.AccessToken,
			WxAccessTokenCreateTime:  t,
			WxRefreshToken:           wxAccessTokenResp.RefreshToken,
			WxRefreshTokenCreateTime: t,
			Avatar:                   wxUserInfoResp.Headimgurl,
			Uid:                      str.GetUid(),
			WxLoginCode:              req.Code,
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
			if _, err := services.NewUserService().UploadAvatar(user, resp.Body, resp.ContentLength, resp.Header.Get("Content-Type")); err != nil {
				log.Println("上传头像失败：", err)
				return
			}
		}(user, wxUserInfoResp.Headimgurl)
	}
	if !user.IsActivated {
		// 邀请码校验
		if len(sliceutil.FilterT(func(code string) bool {
			return req.InviteCode == code
		}, InviteCodeList...)) == 0 {
			//response.BadRequest(c, "邀请码错误")
			response.Resp(c, http.StatusBadRequest, "邀请码错误", map[string]interface{}{"id": user.Id})
			return
		}
		if services.NewInviteCodeService().Create(&models.InviteCode{
			Code:   req.InviteCode,
			UserId: user.Id,
		}) != err {
			log.Println("创建邀请码记录失败：", err)
			//response.BadRequest(c, "邀请码错误")
			response.Resp(c, http.StatusBadRequest, "邀请码错误", map[string]interface{}{"id": str.IntToString(user.Id)})
			return
		}
		// 激活用户
		user.IsActivated = true
		if err := userService.Updates(user); err != nil {
			response.Fail(c, "登陆失败")
			return
		}
	}
	// 创建JWT
	token, err := jwt.CreateJwt(&jwt.Data{
		Id:       str.IntToString(user.Id),
		Nickname: user.Nickname,
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, &wxLoginResp{
		Id:       str.IntToString(user.Id),
		Nickname: user.Nickname,
		Token:    token,
		Avatar:   user.Avatar,
	})
}
