package v1

import (
	"github.com/gin-gonic/gin"
	handlers "kcaitech.com/kcserver/handlers"
)

func loadFeedbackRoutes(api *gin.RouterGroup) {
	//router := api.Group("/")
	router := api
	router.POST("/feedback", handlers.PostFeedback)
}
