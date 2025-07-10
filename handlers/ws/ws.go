package ws

import (
	"log"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/handlers/common"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils/websocket"
)

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

	// 建立ws连接
	ws, err := websocket.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("ws-建立ws连接失败：", userId, err)
		common.ServerError(c, "建立ws连接失败")
		return
	}
	defer ws.Close()

	log.Println("websocket连接成功")

	NewWSClient(ws, token, userId).Serve()
}
