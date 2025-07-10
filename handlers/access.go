package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"kcaitech.com/kcserver/handlers/common"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils/websocket"

	wsclient "kcaitech.com/kcserver/handlers/ws"
)

func AccessGrant(c *gin.Context) {
	var grantPost services.GrantPost
	if err := c.ShouldBindJSON(&grantPost); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	accessKey := uuid.New().String()
	accessSecret := uuid.New().String()

	accessAuthService := services.NewAccessAuthService()
	if err := accessAuthService.UpdateAccessAuth(accessKey, accessSecret, grantPost); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "update access_auth failed"})
		return
	}

	common.Success(c, gin.H{"access_key": accessKey, "access_secret": accessSecret})
}

type AccessUpdatePost struct {
	services.GrantPost
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`
}

func AccessUpdate(c *gin.Context) {
	var updateAccessAuthPost AccessUpdatePost
	if err := c.ShouldBindJSON(&updateAccessAuthPost); err != nil {
		common.BadRequest(c, "invalid request")
		return
	}

	// 验证key是否是当前用户的
	accessAuthService := services.NewAccessAuthService()
	accessAuth, err := accessAuthService.GetAccessAuth(updateAccessAuthPost.AccessKey)
	if err != nil {
		common.BadRequest(c, "access_key or access_secret is invalid")
		return
	}

	if accessAuth.Key != updateAccessAuthPost.AccessKey {
		common.BadRequest(c, "access_key or access_secret is invalid")
		return
	}

	if err := accessAuthService.UpdateAccessAuth(updateAccessAuthPost.AccessKey, updateAccessAuthPost.AccessSecret, updateAccessAuthPost.GrantPost); err != nil {
		common.BadRequest(c, "update access_auth failed")
		return
	}

	if len(updateAccessAuthPost.GrantPost.Document) > 0 {
		accessAuthService.DeleteAccessAuthResourceNotExists(updateAccessAuthPost.AccessKey, uint8(models.AccessAuthResourceTypeDocument), updateAccessAuthPost.GrantPost.Document)
	}
	if len(updateAccessAuthPost.GrantPost.Project) > 0 {
		accessAuthService.DeleteAccessAuthResourceNotExists(updateAccessAuthPost.AccessKey, uint8(models.AccessAuthResourceTypeProject), updateAccessAuthPost.GrantPost.Project)
	}
	if len(updateAccessAuthPost.GrantPost.Team) > 0 {
		accessAuthService.DeleteAccessAuthResourceNotExists(updateAccessAuthPost.AccessKey, uint8(models.AccessAuthResourceTypeTeam), updateAccessAuthPost.GrantPost.Team)
	}

	common.Success(c, gin.H{"message": "update access_auth success"})
}

func AccessToken(c *gin.Context) {
	// 验证access_key, access_secret
	accessKey := c.PostForm("access_key")
	accessSecret := c.PostForm("access_secret")
	if accessKey == "" || accessSecret == "" {
		common.BadRequest(c, "access_key and access_secret are required")
		return
	}

	accessAuthService := services.NewAccessAuthService()
	accessAuth, err := accessAuthService.GetAccessAuth(accessKey)
	if err != nil {
		common.BadRequest(c, "access_key or access_secret is invalid")
		return
	}

	// has access_secret
	if services.HashPassword(accessSecret) != accessAuth.Secret {
		common.BadRequest(c, "access_key or access_secret is invalid")
		return
	}

	// 生成token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"access_key": accessKey,
		"exp":        time.Now().Add(time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(accessAuth.Secret))
	if err != nil {
		common.BadRequest(c, "generate token failed")
		return
	}

	common.Success(c, gin.H{"token": tokenString})
}

// Ws websocket连接
func AccessWs(c *gin.Context) {
	// get token
	token := c.Query("token")
	if token == "" {
		log.Println("ws-未登录")
		common.Unauthorized(c)
		return
	}

	// jwt解析token, 获取access_key
	tokenObj, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return nil, nil
	})
	if err != nil {
		log.Println("ws-Token错误", err)
		common.Unauthorized(c)
		return
	}

	claims, ok := tokenObj.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("ws-Token claims错误")
		common.Unauthorized(c)
		return
	}

	accessKey, ok := claims["access_key"].(string)
	if !ok {
		log.Println("ws-access_key不存在")
		common.Unauthorized(c)
		return
	}

	accessAuthService := services.NewAccessAuthService()
	accessAuth, err := accessAuthService.GetAccessAuth(accessKey)
	if err != nil {
		log.Println("ws-access_key错误", err)
		common.Unauthorized(c)
		return
	}

	// 验证token是否有效
	_, err = jwt.ParseWithClaims(token, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(accessAuth.Secret), nil
	})
	if err != nil {
		log.Println("ws-Token错误", err)
		common.Unauthorized(c)
		return
	}

	userId := accessAuth.UserId

	// 建立ws连接
	ws, err := websocket.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("ws-建立ws连接失败：", userId, err)
		common.ServerError(c, "建立ws连接失败")
		return
	}
	defer ws.Close()

	log.Println("websocket连接成功")

	wsclient.NewWSClient(ws, token, userId).Serve()
}
