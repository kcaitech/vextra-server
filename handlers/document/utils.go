package document

import (
	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/providers/auth"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils"
)

func GetUsersInfo(c *gin.Context, userIds []string) (map[string]*auth.UserInfo, error, int) {
	token, _ := utils.GetAccessToken(c)
	users, err, statusCode := services.GetKCAuthClient().GetUsersInfo(token, userIds)
	if err != nil {
		return nil, err, statusCode
	}

	userMap := make(map[string]*auth.UserInfo)

	// 将用户信息转换为map以便快速查找
	for _, user := range users {
		userMap[user.UserID] = &user
	}
	return userMap, nil, statusCode
}

func GetUserInfo(c *gin.Context) (*auth.UserInfo, error) {
	token, _ := utils.GetAccessToken(c)
	users, err := services.GetKCAuthClient().GetUserInfo(token)
	if err != nil {
		return nil, err
	}
	return users, nil
}
