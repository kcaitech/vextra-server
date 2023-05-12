package controllers

import (
	"encoding/json"
	"fmt"
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
	. "protodesign.cn/kcserver/utils/time"
	"time"
)

type wxLoginReq struct {
	Code string `json:"code" binding:"required"`
}

type wxLoginResp struct {
	Id       int64  `json:"id"`
	Nickname string `json:"nickname"`
	Token    string `json:"token"`
	Avatar   string `json:"avatar"`
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

	userService := services.NewUserService()

	// 创建用户
	user := &models.User{}
	err = userService.Get(user, "wx_open_id = ?", wxAccessTokenResp.Openid)
	if err != nil && nil != services.ErrRecordNotFound {
		log.Println("userService.Get错误：", err)
		response.Fail(c, "登陆失败")
		return
	}
	if err == services.ErrRecordNotFound {
		t := Time(time.Now())
		user = &models.User{
			Nickname:                 wxUserInfoResp.Nickname,
			WxOpenId:                 wxAccessTokenResp.Openid,
			WxAccessToken:            wxAccessTokenResp.AccessToken,
			WxAccessTokenCreateTime:  t,
			WxRefreshToken:           wxAccessTokenResp.RefreshToken,
			WxRefreshTokenCreateTime: t,
			Avatar:                   wxUserInfoResp.Headimgurl,
		}
		err := userService.Create(user)
		if err != nil {
			log.Println(err)
			response.Fail(c, "登陆失败")
			return
		}
	}
	// 创建JWT
	token, err := jwt.CreateJwt(&jwt.Data{
		Id:       fmt.Sprintf("%d", user.Id),
		Nickname: user.Nickname,
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, wxLoginResp{
		Id:       user.Id,
		Nickname: user.Nickname,
		Token:    token,
		Avatar:   user.Avatar,
	})
}
