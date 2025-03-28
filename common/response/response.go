package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func Abort(c *gin.Context, code int, message string, data any) Response {
	resp := Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
	c.AbortWithStatusJSON(http.StatusOK, resp)
	return resp
}

func Resp(c *gin.Context, code int, message string, data any) Response {
	resp := Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
	c.JSON(http.StatusOK, resp)
	return resp
}

func Success(c *gin.Context, data any) Response {
	return Resp(c, 0, "成功", data)
}

func Fail(c *gin.Context, message string) Response {
	if message == "" {
		message = "操作失败"
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

func BadRequestData(c *gin.Context, message string, data any) Response {
	if message == "" {
		message = "参数错误"
	}
	return Resp(c, http.StatusBadRequest, message, data)
}

func Forbidden(c *gin.Context, message string) Response {
	if message == "" {
		message = "无访问权限"
	}
	return Resp(c, http.StatusForbidden, message, nil)
}
