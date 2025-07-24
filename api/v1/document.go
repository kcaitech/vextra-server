package v1

import (
	"github.com/gin-gonic/gin"

	"kcaitech.com/kcserver/handlers/common"
	handlers "kcaitech.com/kcserver/handlers/document"
)

func loadDocumentRoutes(api *gin.RouterGroup) {

	router := api.Group("/documents")

	router.GET("/access_records", handlers.GetUserDocumentAccessRecordsList)    // 获取用户文档访问记录列表
	router.DELETE("/access_record", handlers.DeleteUserDocumentAccessRecord)    // 移除历史记录
	router.GET("/favorites", handlers.GetUserDocumentFavoritesList)             // 获取收藏列表
	router.PUT("/favorites", handlers.SetUserDocumentFavoriteStatus)            // 设置收藏列表
	router.GET("/", handlers.GetUserDocumentList)                               // 获取文档列表
	router.DELETE("/", handlers.DeleteUserDocument)                             // 移动文件到回收站
	router.PUT("/name", handlers.SetDocumentName)                               // 文件重命名
	router.GET("/recycle_bin", handlers.GetUserRecycleBinDocumentList)          // 获取回收站列表
	router.PUT("/recycle_bin", handlers.RestoreUserRecycleBinDocument)          // 恢复文件
	router.DELETE("/recycle_bin", handlers.DeleteUserRecycleBinDocument)        // 彻底删除文件
	router.GET("/info", handlers.GetUserDocumentInfo)                           // 获取文档信息
	router.GET("/permission", handlers.GetUserDocumentPerm)                     // 获取文档权限
	router.GET("/access_key", handlers.GetDocumentAccessKey)                    // 获取文档密钥
	router.POST("/copy", handlers.CopyDocument)                                 // 复制文档
	router.GET("/resource", handlers.GetResourceDocumentList)                   // 获取资源文档列表
	router.POST("/resource", handlers.CreateResourceDocument)                   // 创建资源文档
	router.POST("/review", common.ReReviewDocument)                             // todo: 重新审核文档
	router.GET("/thumbnail_access_key", handlers.GetDocumentThumbnailAccessKey) // 获取文档缩略图
	// 评论
	router.GET("/comments", handlers.GetDocumentComment)         // 获取文档评论
	router.POST("/comment", handlers.PostUserComment)            // 创建评论
	router.PUT("/comment", handlers.PutUserComment)              // 编辑评论
	router.DELETE("/comment", handlers.DeleteUserComment)        // 删除评论
	router.PUT("/comment/status", handlers.SetUserCommentStatus) // 设置评论状态
}
