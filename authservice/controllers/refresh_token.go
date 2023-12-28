package controllers

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/jwt"
	"protodesign.cn/kcserver/utils/str"
	"time"
)

func RefreshToken(c *gin.Context) {
	user, err := auth.GetUser(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	exp, err := auth.GetJwtExp(c)
	if err != nil {
		response.Fail(c, "")
		return
	}
	expRemain := exp - time.Now().Unix()
	if expRemain < 0 {
		response.Fail(c, "token已过期")
		return
	} else if expRemain > 60*60 {
		response.Fail(c, "token未过期")
		return
	}
	token, _ := jwt.CreateJwt(&jwt.Data{
		Id:       str.IntToString(user.Id),
		Nickname: user.Nickname,
	})
	response.Success(c, map[string]any{
		"token": token,
	})
}
