package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	StatusContentReviewFail = 494 // 审核失败
	StatusDocumentNotFound  = 495 // 文档不存在
)

type Response struct {
	Code       int    `json:"code"`
	Message    string `json:"message,omitempty"`
	Data       any    `json:"data,omitempty"`
	Sha1       string `json:"sha1,omitempty"`
	HasMore    bool   `json:"has_more,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
}

func SuccessWithCursor(c *gin.Context, data any, has_more bool, next_cursor string) Response {
	resp := Response{
		Code: http.StatusOK,
		Data: data,
	}

	if has_more {
		resp.HasMore = has_more
		resp.NextCursor = next_cursor
	}

	c.JSON(http.StatusOK, resp)
	return resp
}

func Resp(c *gin.Context, code int, message string, data any) Response {
	resp := Response{
		Code:    code,
		Message: message,
		Data:    data,
	}

	c.JSON(code, resp)
	return resp
}

func Success(c *gin.Context, data any) Response {
	return Resp(c, http.StatusOK, "", data)
}

func Unauthorized(c *gin.Context) Response {
	// if message == "" {
	// 	message = "未登录"
	// }
	return Resp(c, http.StatusUnauthorized, "", nil)
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

func ServerError(c *gin.Context, message string) Response {
	if message == "" {
		message = "服务器出错"
	}
	return Resp(c, http.StatusInternalServerError, message, nil)
}

func ReviewFail(c *gin.Context, message string) Response {
	if message == "" {
		message = "审核失败"
	}
	return Resp(c, StatusContentReviewFail, message, nil)
}
