package http

import (
	"github.com/gin-gonic/gin"

	handlers "kcaitech.com/kcserver/handlers/document"
)

func loadDocumentRoutes(api *gin.RouterGroup) {

	router := api.Group("/documents")

	router.GET("/access_records", handlers.GetUserDocumentAccessRecordsList)
	router.DELETE("/access_record", handlers.DeleteUserDocumentAccessRecord)
	router.GET("/favorites", handlers.GetUserDocumentFavoritesList)
	router.PUT("/favorites", handlers.SetUserDocumentFavoriteStatus)
	router.GET("/", handlers.GetUserDocumentList)
	router.DELETE("/", handlers.DeleteUserDocument) // 移动文件到回收站
	router.PUT("/name", handlers.SetDocumentName)
	router.GET("/recycle_bin", handlers.GetUserRecycleBinDocumentList)
	router.PUT("/recycle_bin", handlers.RestoreUserRecycleBinDocument)
	router.DELETE("/recycle_bin", handlers.DeleteUserRecycleBinDocument)
	router.GET("/info", handlers.GetUserDocumentInfo)        // 获取文档信息
	router.GET("/permission", handlers.GetUserDocumentPerm)  // 获取文档权限
	router.GET("/access_key", handlers.GetDocumentAccessKey) // 获取文档密钥
	router.POST("/copy", handlers.CopyDocument)
	router.GET("/resource", handlers.GetResourceDocumentList)
	router.POST("/resource", handlers.CreateResourceDocument)
	// 评论
	router.GET("/comments", handlers.GetDocumentComment)         // 获取文档评论
	router.POST("/comment", handlers.PostUserComment)            // 创建评论
	router.PUT("/comment", handlers.PutUserComment)              // 编辑评论
	router.DELETE("/comment", handlers.DeleteUserComment)        // 删除评论
	router.PUT("/comment/status", handlers.SetUserCommentStatus) // 设置评论状态
}
