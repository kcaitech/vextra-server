package v1

import (
	"github.com/gin-gonic/gin"
	ws "kcaitech.com/kcserver/handlers/ws"
)

func loadWsRoutes(api *gin.RouterGroup) {
	//router := api.Group("/")
	router := api
	router.GET("/ws", ws.Ws)
}
