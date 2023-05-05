package http

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/documentservice/controllers"
)

func loadDocumentRoutes(api *gin.RouterGroup) {
	router := api.Group("/documents")
	router.GET("/upload", controllers.UploadHandler)
	router.GET("/", controllers.DocumentUserList)
	router.GET("/access_records", controllers.DocumentUserAccessRecordsList)
}
