package auth

import (
	"github.com/gin-gonic/gin"
	. "protodesign.cn/kcserver/common/jwt"
	"protodesign.cn/kcserver/utils/strutil"
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
	id, err := strutil.ToInt(jwtData.Id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
