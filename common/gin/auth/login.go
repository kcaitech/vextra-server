package auth

import (
	"github.com/gin-gonic/gin"
	. "kcaitech.com/kcserver/common/jwt"
	"kcaitech.com/kcserver/common/models"
	"kcaitech.com/kcserver/common/services"
	"kcaitech.com/kcserver/utils/str"
)

func GetJwt(c *gin.Context) string {
	return GetJwtFromAuthorization(c.GetHeader("Authorization"))
}

func GetJwtData(c *gin.Context) (*Data, error) {
	return ParseJwt(GetJwt(c))
}

func GetJwtExp(c *gin.Context) (int64, error) {
	return ParseJwtExp(GetJwt(c))
}

func GetJwtNbf(c *gin.Context) (int64, error) {
	return ParseJwtNbf(GetJwt(c))
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
