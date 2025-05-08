package utils

import (
	"errors"
	"strconv"

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

// QueryInt 从请求的查询参数中获取整数值，如果参数不存在或无法转换为整数，则返回默认值
func QueryInt(c *gin.Context, key string, defaultVal int) int {
	strVal := c.Query(key)
	if strVal == "" {
		return defaultVal
	}

	intVal, err := strconv.Atoi(strVal)
	if err != nil {
		return defaultVal
	}

	return intVal
}
