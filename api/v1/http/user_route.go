package http

import (
	"encoding/base64"
	"log"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/common/response"
	"kcaitech.com/kcserver/providers/auth"
	safereviewBase "kcaitech.com/kcserver/providers/safereview"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils"

	// "kcaitech.com/kcserver/common"
	// . "kcaitech.com/kcserver/common/gin/reverse_proxy"
	controllers "kcaitech.com/kcserver/controllers/user"
)

func loadUserRoutes(api *gin.RouterGroup) {
	router := api.Group("/users")

	authorized := router.Group("/")
	// authorized.Use(services.GetKCAuthClient().AuthRequired())
	{
		authorized.GET("/info", GetUserInfo)
		authorized.PUT("/info/nickname", SetNickname)
		authorized.PUT("/info/avatar", SetAvatar)
		authorized.GET("/kv_storage", controllers.GetUserKVStorage)
		authorized.POST("/kv_storage", controllers.SetUserKVStorage)
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
		response.ServerError(c, "操作失败")
		return
	}

	response.Success(c, map[string]any{
		"id":       user.UserID,
		"nickname": user.Profile.Nickname,
		"avatar":   user.Profile.Avatar,
	})
}
func SetNickname(c *gin.Context) {
	var req struct {
		Nickname string `json:"nickname" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	// 审核
	safereview := services.GetSafereviewClient()
	if safereview != nil {
		reviewResponse, err := safereview.ReviewText(req.Nickname)
		if err != nil || reviewResponse.Status != safereviewBase.ReviewTextResultPass {
			log.Println("昵称审核不通过", req.Nickname, err, reviewResponse)
			response.ReviewFail(c, "审核不通过")
			return
		}
	}
	// get user info
	user, err := get_user_info(c)
	if err != nil {
		response.ServerError(c, "操作失败")
		return
	}
	user.Profile.Nickname = req.Nickname
	client := services.GetKCAuthClient()
	token, err := utils.GetAccessToken(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	err = client.UpdateUserInfo(token, user)
	if err != nil {
		response.ServerError(c, "操作失败")
		return
	}
	response.Success(c, "")
}
func SetAvatar(c *gin.Context) {
	// user, err := auth.GetUser(c)
	// if err != nil {
	// 	response.Unauthorized(c)
	// 	return
	// }
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "参数错误：file")
		return
	}
	if fileHeader.Size > 2<<20 {
		response.BadRequest(c, "文件大小不能超过2MB")
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		response.BadRequest(c, "获取文件失败")
		return
	}
	defer file.Close()
	fileBytes := make([]byte, fileHeader.Size)
	if _, err := file.Read(fileBytes); err != nil {
		response.BadRequest(c, "读取文件失败")
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
			response.ReviewFail(c, "头像审核失败")
			return

		} else if reviewResponse.Status != safereviewBase.ReviewImageResultPass {
			log.Println("头像审核不通过", err, reviewResponse)
			response.ReviewFail(c, "头像审核不通过")
			return
		}
	}
	token, err := utils.GetAccessToken(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	client := services.GetKCAuthClient()
	err = client.UpdateAvatar(token, fileBytes, fileHeader.Filename)
	if err != nil {
		response.ServerError(c, "操作失败")
		return
	}
	response.Success(c, "")
}
