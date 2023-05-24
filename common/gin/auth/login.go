package auth

import (
	"github.com/gin-gonic/gin"
	. "protodesign.cn/kcserver/common/jwt"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/utils/str"
)

func GetJwtData(c *gin.Context) (*Data, error) {
	token := GetJwtFromAuthorization(c.GetHeader("Authorization"))
	return ParseJwt(token)
}

func GetUserId(c *gin.Context) (int64, error) {
	jwtData, err := GetJwtData(c)
	if err != nil {
		return 0, err
	}
	id, err := str.ToInt(jwtData.Id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetUser(c *gin.Context) (*models.User, error) {
	userId, err := GetUserId(c)
	if err != nil {
		return nil, err
	}
	var user models.User
	if err := services.NewUserService().GetById(userId, &user); err != nil {
		return nil, err
	}
	return &user, nil
}
