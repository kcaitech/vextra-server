package controllers

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/utils/sliceutil"
)

var AllowedKeyList = []string{
	"FontList",
}

func GetUserKVStorage(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}

	key := c.Query("key")
	if key == "" {
		response.BadRequest(c, "参数错误：key")
		return
	}

	if len(sliceutil.FilterT(func(code string) bool {
		return key == code
	}, AllowedKeyList...)) == 0 {
		response.BadRequest(c, "不允许的key: "+key)
		return
	}

	result := map[string]any{}
	userKVStorageService := services.NewUserKVStorageService()
	if userKVStorage, err := userKVStorageService.GetOne(userId, key); err == nil {
		result[key] = userKVStorage
	}

	response.Success(c, result)
}

func SetUserKVStorage(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.Key == "" {
		response.BadRequest(c, "参数错误：key")
		return
	}

	userKVStorageService := services.NewUserKVStorageService()
	if !userKVStorageService.SetOne(userId, req.Key, req.Value) {
		response.Fail(c, "操作失败")
		return
	}

	response.Success(c, "")
}
