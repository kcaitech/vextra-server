package http

import (
	"encoding/base64"
	"log"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/providers/auth"
	safereviewBase "kcaitech.com/kcserver/providers/safereview"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils"

	// "kcaitech.com/kcserver/common"
	// . "kcaitech.com/kcserver/common/gin/reverse_proxy"
	"kcaitech.com/kcserver/handlers/common"
	handlers "kcaitech.com/kcserver/handlers/user"
)

func loadUserRoutes(api *gin.RouterGroup) {
	router := api.Group("/users")

	authorized := router.Group("/")
	// authorized.Use(services.GetKCAuthClient().AuthRequired())
	{
		authorized.GET("/info", GetUserInfo)
		authorized.PUT("/info/nickname", SetNickname)
		authorized.PUT("/info/avatar", SetAvatar)
		authorized.GET("/kv_storage", handlers.GetUserKVStorage)
		authorized.POST("/kv_storage", handlers.SetUserKVStorage)
	}
}

func get_user_info(c *gin.Context) (*auth.UserInfo, error) {
	token, err := utils.GetAccessToken(c)
	if err != nil {
		return nil, err
	}
	client := services.GetKCAuthClient()
	// if userId == "" {
	return client.GetUserInfo(token)
	// } else {
	// return client.GetUserInfoById(token, userId)
	// }
}

func GetUserInfo(c *gin.Context) {
	// userId := c.Query("user_id")
	user, err := get_user_info(c)
	if err != nil {
		log.Println("获取用户信息失败", err)
		common.ServerError(c, "操作失败")
		return
	}

	common.Success(c, map[string]any{
		"id":       user.UserID,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
	})
}
func SetNickname(c *gin.Context) {
	var req struct {
		Nickname string `json:"nickname" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, err.Error())
		return
	}
	// 审核
	safereview := services.GetSafereviewClient()
	if safereview != nil {
		reviewResponse, err := safereview.ReviewText(req.Nickname)
		if err != nil {
			log.Println("昵称审核失败", req.Nickname, err)
			common.ReviewFail(c, "审核失败")
			return
		} else if reviewResponse != nil && reviewResponse.Status != safereviewBase.ReviewTextResultPass {
			log.Println("昵称审核不通过", req.Nickname, reviewResponse)
			common.ReviewFail(c, "审核不通过")
			return
		}
	}
	// get user info
	user, err := get_user_info(c)
	if err != nil {
		common.ServerError(c, "操作失败")
		return
	}
	user.Nickname = req.Nickname
	client := services.GetKCAuthClient()
	token, err := utils.GetAccessToken(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	err = client.UpdateUserInfo(token, user)
	if err != nil {
		common.ServerError(c, "操作失败")
		return
	}
	common.Success(c, "")
}
func SetAvatar(c *gin.Context) {
	// user, err := auth.GetUser(c)
	// if err != nil {
	// 	common.Unauthorized(c)
	// 	return
	// }
	fileHeader, err := c.FormFile("file")
	if err != nil {
		common.BadRequest(c, "参数错误：file")
		return
	}
	if fileHeader.Size > 2<<20 {
		common.BadRequest(c, "文件大小不能超过2MB")
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		common.BadRequest(c, "获取文件失败")
		return
	}
	defer file.Close()
	fileBytes := make([]byte, fileHeader.Size)
	if _, err := file.Read(fileBytes); err != nil {
		common.BadRequest(c, "读取文件失败")
		return
	}
	// contentType := fileHeader.Header.Get("Content-Type")
	base64Str := base64.StdEncoding.EncodeToString(fileBytes)
	// 审核
	safereview := services.GetSafereviewClient()
	if safereview != nil {
		reviewResponse, err := safereview.ReviewPictureFromBase64(base64Str)
		if err != nil {
			log.Println("头像审核失败", err)
			common.ReviewFail(c, "头像审核失败")
			return

		} else if reviewResponse.Status != safereviewBase.ReviewImageResultPass {
			log.Println("头像审核不通过", err, reviewResponse)
			common.ReviewFail(c, "头像审核不通过")
			return
		}
	}
	token, err := utils.GetAccessToken(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	client := services.GetKCAuthClient()
	url, err := client.UpdateAvatar(token, fileBytes, fileHeader.Filename)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}
	common.Success(c, map[string]any{
		"avatar": url,
	})
}
