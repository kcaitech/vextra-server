package document

import (
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/providers/auth"
	"kcaitech.com/kcserver/providers/storage"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils"
)

func GetUsersInfo(c *gin.Context, userIds []string) (map[string]*auth.UserInfo, error) {
	token, _ := utils.GetAccessToken(c)
	users, err := services.GetKCAuthClient().GetUsersInfo(token, userIds)
	if err != nil {
		return nil, err
	}

	userMap := make(map[string]*auth.UserInfo)

	// 将用户信息转换为map以便快速查找
	for _, user := range users {
		userMap[user.UserID] = &user
	}
	return userMap, nil
}

func GetUserInfo(c *gin.Context) (*auth.UserInfo, error) {
	token, _ := utils.GetAccessToken(c)
	users, err := services.GetKCAuthClient().GetUserInfo(token)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func GetDocumentThumbnail(c *gin.Context, documentId string, storage *storage.StorageClient) string {
	objects := storage.Bucket.ListObjects(documentId + "/thumbnail/")
	for object := range objects {
		if object.Err == nil {
			// 生成预签名URL，有效期1小时
			reqParams := make(url.Values)
			reqParams.Set("response-content-disposition", "inline")
			presignedURL, err := storage.Bucket.PresignedGetObject(object.Key, time.Hour, nil)
			if err == nil {
				return presignedURL
			}
			break
		}
	}
	return ""
}
