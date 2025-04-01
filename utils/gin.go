package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func GetUserId(c *gin.Context) (string, error) {
	_userId, _ := c.Get("user_id")
	if _userId == nil {
		return "", errors.New("用户未登录")
	}
	return _userId.(string), nil
}

func GetAccessToken(c *gin.Context) (string, error) {
	token, _ := c.Get("access_token")
	if token == nil {
		return "", errors.New("用户未登录")
	}
	return token.(string), nil
}
