package controllers

import (
	"encoding/json"
	"errors"
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

type wxOpenWebLoginReq struct {
	Id         string `json:"id"`
	Code       string `json:"code" binding:"required"`
	InviteCode string `json:"invite_code"`
}

var InviteCodeList = []string{ // 邀请码
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
	"i41OOyAU",
	"60EH3F2V",
	"Arls34gB",
	"BS8uQsNv",
	"GDwUUopZ",
	"crp3bk0m",
	"7NA2YWVJ",
	"dz8IyaJU",
	"OT7Z8Ok2",
	"9ydop6A1",
	"locZ7h4f",
	"c93vFuU2",
	"0SWBX2Tu",
	"cD9tU1ik",
	"doDo74pK",
	"1wPkjY9r",
	"8QksT8js",
	"4DyBu91W",
	"6gHNhhn2",
	"IqaebwG6",
	"4q0raoz2",
	"56X8Nzki",
	"sifAHR9y",
	"T55chxqa",
	"NAOOJltS",
	"T3d8Ag41",
	"wSIkgMz3",
	"DEB2sTzl",
	"lf5Z5gkD",
	"yAr8XWhx",
	"h9mawoXV",
	"TX2pCZsn",
	"0JSDy3J8",
	"SIO27tC7",
	"M2Iwm8nt",
	"MxX3P70R",
	"3Cf7p0si",
	"NhznNy4P",
	"98XH4Ad5",
	"6je4srgM",
	"1G9YNZ6H",
	"0jD2Sr8m",
	"Y07tK1Sk",
	"3gE88z6p",
	"XCELvMuL",
	"Mc0MbRd3",
	"zUoCurpT",
	"8zX5G223",
	"12Rc0E8V",
	"VqwHcP1F",
	"3hgCKnk3",
	"kU2NQNvx",
	"c6X1psV5",
	"h6e6c487",
	"gMy0py6e",
	"95sg47Bm",
	"5eA4GmCu",
	"ShW7GRnb",
	"U9k6S9UK",
	"yTCV316u",
	"4T78lRBo",
	"pNa3GQGk",
	"E5EMV97P",
	"Qzf8ip5L",
	"uQ97y6jk",
	"lErxvwJa",
	"w2Cone3v",
	"c7Z68r05",
	"3LC1dJE9",
	"thH13Zc5",
	"9y2WLm4T",
	"qu5Bl47U",
	"9HzljAw4",
	"1eVYaSm7",
	"0By53S7C",
	"bxfY22Bh",
	"7VVBR0R5",
	"goNEU1rm",
	"Hf1yFR3V",
	"jRxHa5i0",
	"wWLe6Jns",
	"gdj4wwu6",
	"MI0vZLjh",
	"b95VXLFG",
	"u16fnVfz",
	"in9iZnUG",
	"MfMV8AyY",
	"b1IZGW69",
	"G2kcS9S8",
	"I9EjetnU",
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
		resp.Avatar = common.FileStorageHost + resp.Avatar
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
			_, _ = userService.Updates(user)
		}
	}

	if IsInviteCodeCheck && !user.IsActivated {
		user.WxLoginCode = req.Code
		_, _ = userService.Updates(user)
		// 邀请码校验
		if len(sliceutil.FilterT(func(code string) bool {
			return req.InviteCode == code
		}, InviteCodeList...)) == 0 {
			//response.BadRequest(c, "邀请码错误")
			response.Resp(c, http.StatusBadRequest, "邀请码错误", map[string]interface{}{"id": str.IntToString(user.Id)})
			return
		}
		if err := services.NewInviteCodeService().Create(&models.InviteCode{
			Code:   req.InviteCode,
			UserId: user.Id,
		}); err != nil {
			log.Println("创建邀请码记录失败：", err)
			//response.BadRequest(c, "邀请码错误")
			response.Resp(c, http.StatusBadRequest, "邀请码错误", map[string]interface{}{"id": str.IntToString(user.Id)})
			return
		}
		// 激活用户
		user.IsActivated = true
		if _, err := userService.Updates(user); err != nil {
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

	response.Success(c, &wxOpenWebLoginResp{
		Id:       str.IntToString(user.Id),
		Nickname: user.Nickname,
		Token:    token,
		Avatar:   user.Avatar,
	})
}
