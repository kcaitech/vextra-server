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
	"kcaitech.com/kcserver/utils"
	"kcaitech.com/kcserver/utils/websocket"

	wsclient "kcaitech.com/kcserver/handlers/ws"
)

func AccessCreate(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}

	var grantPost services.GrantPost
	if err := c.ShouldBindJSON(&grantPost); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	accessKey := uuid.New().String()
	accessSecret := uuid.New().String()

	accessAuthService := services.NewAccessAuthService()
	if err := accessAuthService.UpdateAccessAuth(userId, accessKey, accessSecret, grantPost); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "update access_auth failed"})
		return
	}

	common.Success(c, gin.H{"access_key": accessKey, "access_secret": accessSecret})
}

func AccessList(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}

	accessAuthService := services.NewAccessAuthService()
	accessAuths, err := accessAuthService.GetCombinedAccessAuth(userId)
	if err != nil {
		common.BadRequest(c, "get access_auth failed")
		return
	}

	common.Success(c, accessAuths)
}

type AccessUpdatePost struct {
	services.GrantPost
	AccessKey string `json:"access_key"`
	// AccessSecret string `json:"access_secret"`
}

func AccessUpdate(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}

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

	if accessAuth.UserId != userId {
		common.BadRequest(c, "access_key is invalid")
		return
	}

	if err := accessAuthService.UpdateAccessAuth(userId, updateAccessAuthPost.AccessKey, "", updateAccessAuthPost.GrantPost); err != nil {
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

type AccessDeletePost struct {
	AccessKey string `json:"access_key"`
}

func AccessDelete(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}

	var accessDeletePost AccessDeletePost
	if err := c.ShouldBindJSON(&accessDeletePost); err != nil {
		common.BadRequest(c, "invalid request")
		return
	}

	if accessDeletePost.AccessKey == "" {
		common.BadRequest(c, "access_key is required")
		return
	}

	accessAuthService := services.NewAccessAuthService()
	accessAuth, err := accessAuthService.GetAccessAuth(accessDeletePost.AccessKey)
	if err != nil {
		common.BadRequest(c, "access_key is invalid")
		return
	}

	if accessAuth.UserId != userId {
		common.BadRequest(c, "access_key is invalid")
		return
	}

	if err := accessAuthService.DeleteAccessAuth(accessDeletePost.AccessKey); err != nil {
		common.BadRequest(c, "delete access_auth failed")
		return
	}

	common.Success(c, gin.H{"message": "delete access_auth success"})
}

type AccessTokenPost struct {
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`
	Expire       *int   `json:"expire,omitempty"`
}

func AccessToken(c *gin.Context) {
	// 验证access_key, access_secret
	var accessTokenPost AccessTokenPost
	if err := c.ShouldBindJSON(&accessTokenPost); err != nil {
		common.BadRequest(c, "invalid request")
		return
	}

	expireInt := int(time.Hour.Seconds()) // 默认值
	if accessTokenPost.Expire != nil {
		expireInt = *accessTokenPost.Expire
	}

	if accessTokenPost.AccessKey == "" || accessTokenPost.AccessSecret == "" {
		common.BadRequest(c, "access_key and access_secret are required")
		return
	}

	accessAuthService := services.NewAccessAuthService()
	accessAuth, err := accessAuthService.GetAccessAuth(accessTokenPost.AccessKey)
	if err != nil {
		log.Println("access_key is invalid", err, accessTokenPost.AccessKey)
		common.BadRequest(c, "access_key is invalid")
		return
	}

	// has access_secret
	if err := services.CheckPassword(accessAuth.Secret, accessTokenPost.AccessSecret); err != nil {
		log.Println("access_secret is invalid", accessTokenPost.AccessSecret, accessAuth.Secret)
		common.BadRequest(c, "access_secret is invalid")
		return
	}

	// 生成token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"access_key": accessTokenPost.AccessKey,
		"exp":        time.Now().Add(time.Duration(expireInt) * time.Second).Unix(),
	})
	tokenString, err := token.SignedString([]byte(accessAuth.Secret))
	if err != nil {
		common.BadRequest(c, "generate token failed")
		return
	}

	common.Success(c, gin.H{"token": tokenString, "expire": expireInt})
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

	tokenObj, _ := jwt.ParseWithClaims(token, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return nil, nil
	})
	claims, ok := tokenObj.Claims.(*jwt.MapClaims)
	if !ok {
		log.Println("ws-Token claims错误")
		common.Unauthorized(c)
		return
	}

	accessKey, ok := (*claims)["access_key"].(string)
	if !ok {
		log.Println("ws-access_key不存在")
		common.Unauthorized(c)
		return
	}

	// 验证token是否过期
	exp, ok := (*claims)["exp"].(float64)
	if !ok || time.Now().Unix() > int64(exp) {
		log.Println("ws-Token过期")
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
		log.Println("jwt parse with claims error", err)
		common.Unauthorized(c)
		return
	}

	defer_func, err := wsclient.CheckConcurrentDocumentLimit()
	if defer_func != nil {
		defer defer_func()
	}
	if err != nil {
		log.Println("ws-并发限制", err)
		common.ServerError(c, "并发限制")
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
