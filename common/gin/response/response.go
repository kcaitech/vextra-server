package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Resp(c *gin.Context, code int, message string, data interface{}) Response {
	resp := Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
	c.JSON(http.StatusOK, resp)
	return resp
}

func Success(c *gin.Context, data interface{}) Response {
	return Resp(c, 0, "成功", data)
}

func Fail(c *gin.Context, message string) Response {
	if message == "" {
		message = "失败"
	}
	return Resp(c, -1, message, nil)
}

func Unauthorized(c *gin.Context) Response {
	return Resp(c, http.StatusUnauthorized, "未登录", nil)
}

func BadRequest(c *gin.Context, message string) Response {
	if message == "" {
		message = "参数错误"
	}
	return Resp(c, http.StatusBadRequest, message, nil)
}

func Forbidden(c *gin.Context, message string) Response {
	if message == "" {
		message = "无访问权限"
	}
	return Resp(c, http.StatusForbidden, message, nil)
}
