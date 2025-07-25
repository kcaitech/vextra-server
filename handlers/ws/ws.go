package ws

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	com "kcaitech.com/kcserver/common"
	"kcaitech.com/kcserver/config"
	"kcaitech.com/kcserver/handlers/common"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils/websocket"
)

func CheckConcurrentDocumentLimit() (defer_func func(), err error) {
	if config.ConcurrentDocumentLimit == 0 {
		return nil, nil
	}

	redisClient := services.GetRedisDB().Client
	connectionKey := fmt.Sprintf("%s%s", com.RedisKeyDocumentConcurrentLimit, uuid.New().String())

	err = redisClient.SetEx(context.Background(), connectionKey, 0, time.Hour*24).Err()
	if err != nil {
		log.Println("ws-设置连接标识失败", err)
		return nil, err
	}

	defer_func = func() {
		redisClient.Del(context.Background(), connectionKey)
	}

	// 统计当前连接数
	keyPattern := com.RedisKeyDocumentConcurrentLimit + ":*"
	keys, err := redisClient.Keys(context.Background(), keyPattern).Result()
	if err != nil {
		log.Println("ws-获取并发连接数失败", err)
		return defer_func, err
	}

	currentCount := int64(len(keys))
	if currentCount > config.ConcurrentDocumentLimit {
		log.Println("ws-并发限制", currentCount)
		return defer_func, errors.New("并发限制")
	}

	return defer_func, nil
}

// Ws websocket连接
func Ws(c *gin.Context) {
	// get token
	token := c.Query("token")
	if token == "" {
		log.Println("ws-未登录")
		common.Unauthorized(c)
		return
	}

	jwtClient := services.GetKCAuthClient()
	claims, err := jwtClient.ValidateToken(token)
	if err != nil {
		log.Println("ws-Token错误", err)
		common.Unauthorized(c)
		return
	}

	userId := claims.UserID
	if userId == "" {
		log.Println("ws-UserId错误", userId)
		common.BadRequest(c, "UserId错误")
		return
	}

	defer_func, err := CheckConcurrentDocumentLimit()
	if defer_func != nil {
		defer defer_func()
	}
	if err != nil {
		log.Println("ws-并发限制", err)
		common.ServerError(c, "并发限制")
		return
	}

	// 建立ws连接
	ws, err := websocket.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("ws-建立ws连接失败：", userId, err)
		common.ServerError(c, "建立ws连接失败")
		return
	}
	defer ws.Close()

	log.Println("websocket连接成功")

	NewWSClient(ws, token, userId, false).Serve()
}
