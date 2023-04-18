package controllers

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/loginservice/models"
	"protodesign.cn/kcserver/loginservice/services"
	"protodesign.cn/kcserver/loginservice/utils/jwt"
	"protodesign.cn/kcserver/utils/gin/response"
)

type loginRequest struct {
	Nickname string `json:"nickname" binding:"required"`
}

type loginResponse struct {
	Id       uint   `json:"id"`
	Nickname string `json:"nickname"`
	WxOpenId string `json:"wx_open_id"`
	Token    string `json:"token"`
}

// Login 登录
func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	userService := services.NewUserService()
	user, _ := userService.GetUserByNickname(req.Nickname)
	if user == nil {
		user = &models.User{
			Nickname: req.Nickname,
			WxOpenId: req.Nickname,
		}
		err := userService.CreateUser(user)
		if err != nil {
			response.Fail(c, err.Error())
			return
		}
	}

	token, err := jwt.CreateJwt(&jwt.Data{
		Id:       user.ID,
		Nickname: user.Nickname,
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, loginResponse{
		Id:       user.ID,
		Nickname: user.Nickname,
		WxOpenId: user.Nickname,
		Token:    token,
	})
}
