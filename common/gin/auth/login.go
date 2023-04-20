package auth

import (
	"github.com/gin-gonic/gin"
	. "protodesign.cn/kcserver/common/jwt"
)

func GetJwtData(c *gin.Context) (*Data, error) {
	token := c.GetHeader("Token")
	return ParseJwt(token)
}

func GetUserId(c *gin.Context) (uint, error) {
	jwtData, err := GetJwtData(c)
	if err != nil {
		return 0, err
	}
	return jwtData.Id, nil
}
