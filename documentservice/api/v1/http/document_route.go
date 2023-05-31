package http

import (
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/documentservice/controllers"
)

func loadDocumentRoutes(api *gin.RouterGroup) {
	router := api.Group("/documents")
	{
		router.GET("/upload", controllers.UploadDocumentByUser)
		router.GET("/ws", controllers.IncrementalUpload)
		router.GET("/access_records", controllers.GetUserDocumentAccessRecordsList)
		router.DELETE("/access_record", controllers.DeleteUserDocumentAccessRecord)
		router.GET("/favorites", controllers.GetUserDocumentFavoritesList)
		router.PUT("/favorites", controllers.SetUserDocumentFavoriteStatus)
		router.GET("/", controllers.GetUserDocumentList)
		router.DELETE("/", controllers.DeleteUserDocument)
		router.PUT("/name", controllers.SetDocumentName)
		router.GET("/shares", controllers.GetUserDocumentSharesList)
		router.DELETE("/share", controllers.DeleteUserShare)
		router.GET("/recycle_bin", controllers.GetUserRecycleBinDocumentList)
		router.PUT("/recycle_bin", controllers.RestoreUserRecycleBinDocument)
		router.DELETE("/recycle_bin", controllers.DeleteUserRecycleBinDocument)
		router.GET("/info", controllers.GetUserDocumentInfo)
		router.GET("/permission", controllers.GetUserDocumentPerm)
		router.GET("/access_key", controllers.GetDocumentAccessKey)
		router.PUT("/shares/set", controllers.SetDocumentShareType)
		router.GET("/shares/all", controllers.GetDocumentSharesList)
		router.PUT("/shares", controllers.SetDocumentSharePermission)
		router.DELETE("/shares", controllers.DeleteDocumentSharePermission)
		router.POST("/shares/apply", controllers.ApplyDocumentPermission)
		router.GET("/shares/apply", controllers.GetDocumentPermissionRequestsList)
		router.POST("/shares/apply/audit", controllers.ReviewDocumentPermissionRequest)
	}
}
