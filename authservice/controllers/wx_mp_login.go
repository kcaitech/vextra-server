package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"net/url"
	"protodesign.cn/kcserver/authservice/config"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/jwt"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/utils/str"
	. "protodesign.cn/kcserver/utils/time"
	"time"
)

type wxMpLoginReq struct {
	Id   string `json:"id"`
	Code string `json:"code" binding:"required"`
}

type wxMpLoginResp struct {
	Id       string `json:"id"`
	Nickname string `json:"nickname"`
	Token    string `json:"token"`
	Avatar   string `json:"avatar"`
}

type wxMpAccessTokenResp struct {
	SessionKey string `json:"session_key"`
	UnionId    string `json:"unionid"`
	ErrMsg     string `json:"errmsg"`
	Openid     string `json:"openid"`
	ErrCode    int32  `json:"errcode"`
}

type wxMpUserInfoResp struct {
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

// WxMpLogin 微信小程序登录
func WxMpLogin(c *gin.Context) {
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
		queryParams.Set("appid", config.Config.WxMp.Appid)
		queryParams.Set("secret", config.Config.WxMp.Secret)
		queryParams.Set("js_code", req.Code)
		queryParams.Set("grant_type", "authorization_code")
		resp, err := http.Get("https://api.weixin.qq.com/sns/jscode2session" + "?" + queryParams.Encode())
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
		var wxAccessTokenResp wxMpAccessTokenResp
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

		err = userService.Get(user, "wx_union_id = ?", wxAccessTokenResp.UnionId)
		if err != nil && !errors.Is(err, services.ErrRecordNotFound) {
			log.Println("userService.Get错误：", err)
			response.Fail(c, "登陆失败")
			return
		}
		if errors.Is(err, services.ErrRecordNotFound) {
			// 创建用户
			t := Time(time.Now())
			// 用sha256加密UnionId，取后12位，前面加"wx_"，后面加4位随机字符，作为昵称
			secret := "123456"
			h := hmac.New(sha256.New, []byte(secret))
			h.Write([]byte(wxAccessTokenResp.UnionId))
			s := base64.StdEncoding.EncodeToString(h.Sum(nil))
			nickname := "wx_" + string(s[len(s)-12:]) + str.GetRandomAlphaStr(4)
			user = &models.User{
				Nickname:                 nickname,
				WxMpOpenId:               wxAccessTokenResp.Openid,
				WxUnionId:                wxAccessTokenResp.UnionId,
				WxMpSessionKey:           wxAccessTokenResp.SessionKey,
				WxMpSessionKeyCreateTime: t,
				WxMpLoginCode:            req.Code,
				Uid:                      str.GetUid(),
			}
			err := userService.Create(user)
			if err != nil {
				log.Println(err)
				response.Fail(c, "登陆失败")
				return
			}
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
