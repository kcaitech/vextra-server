package http

import (
	"github.com/gin-gonic/gin"

	handlers "kcaitech.com/kcserver/handlers/document"
)

func loadDocumentRoutes(api *gin.RouterGroup) {

	router := api.Group("/documents")

	// 不验证登录
	// router.GET("/shares/wx_mp_code", handlers.GetWxMpCode)

	// 登陆验证
	// router.Use(services.GetKCAuthClient().AuthRequired())

	router.GET("/access_records", handlers.GetUserDocumentAccessRecordsList)
	router.DELETE("/access_record", handlers.DeleteUserDocumentAccessRecord)
	router.GET("/favorites", handlers.GetUserDocumentFavoritesList)
	router.PUT("/favorites", handlers.SetUserDocumentFavoriteStatus)
	router.GET("/", handlers.GetUserDocumentList)
	router.DELETE("/", handlers.DeleteUserDocument)
	router.PUT("/name", handlers.SetDocumentName)
	router.GET("/recycle_bin", handlers.GetUserRecycleBinDocumentList)
	router.PUT("/recycle_bin", handlers.RestoreUserRecycleBinDocument)
	router.DELETE("/recycle_bin", handlers.DeleteUserRecycleBinDocument)
	router.GET("/info", handlers.GetUserDocumentInfo)
	router.GET("/permission", handlers.GetUserDocumentPerm)
	router.GET("/access_key", handlers.GetDocumentAccessKey)
	// 复制文档
	router.POST("/copy", handlers.CopyDocument)
	// 分享
	router.GET("/shares/receives", handlers.GetUserReceiveSharesList) // 查询用户加入的文档分享列表
	router.DELETE("/share", handlers.DeleteUserShare)
	router.PUT("/shares/set", handlers.SetDocumentShareType)
	router.GET("/shares/grantees", handlers.GetDocumentSharesList) // 查询某个文档对所有用户的分享列表
	router.PUT("/shares", handlers.SetDocumentSharePermission)
	router.DELETE("/shares", handlers.DeleteDocumentSharePermission)
	router.POST("/shares/apply", handlers.ApplyDocumentPermission)
	router.GET("/shares/apply", handlers.GetDocumentPermissionRequestsList)
	router.POST("/shares/apply/audit", handlers.ReviewDocumentPermissionRequest)
	// 评论
	router.GET("/comments", handlers.GetDocumentComment)
	router.POST("/comment", handlers.PostUserComment)
	router.PUT("/comment", handlers.PutUserComment)
	router.DELETE("/comment", handlers.DeleteUserComment)
	router.PUT("/comment/status", handlers.SetUserCommentStatus)
	// team
	router.POST("/team", handlers.CreateTeam)
	router.GET("/team/list", handlers.GetTeamList)
	router.GET("/team/member/list", handlers.GetTeamMemberList)
	router.DELETE("/team", handlers.DeleteTeam)
	router.POST("/team/apply", handlers.ApplyJoinTeam)
	router.GET("/team/apply", handlers.GetTeamJoinRequestList)
	router.GET("/team/self_apply", handlers.GetSelfTeamJoinRequestList)
	router.POST("/team/apply/audit", handlers.ReviewTeamJoinRequest)
	router.PUT("/team/info", handlers.SetTeamInfo)
	router.PUT("/team/invited", handlers.SetTeamInvited)
	router.GET("/team/info/invited", handlers.GetTeamInvitedInfo)
	router.POST("/team/exit", handlers.ExitTeam)
	router.PUT("/team/member/perm", handlers.SetTeamMemberPermission)
	router.PUT("/team/creator", handlers.ChangeTeamCreator)
	router.DELETE("/team/member", handlers.RemoveTeamMember)
	router.PUT("/team/member/nickname", handlers.SetTeamMemberNickname)
	// team/project
	router.POST("/team/project", handlers.CreateProject)
	router.GET("/team/project/list", handlers.GetProjectList)
	router.GET("/team/project/member/list", handlers.GetProjectMemberList)
	router.DELETE("/team/project", handlers.DeleteProject)
	router.POST("/team/project/apply", handlers.ApplyJoinProject)
	router.GET("/team/project/apply", handlers.GetProjectJoinRequestList)
	router.GET("/team/project/self_apply", handlers.GetSelfProjectJoinRequestList)
	router.POST("/team/project/apply/audit", handlers.ReviewProjectJoinRequest)
	router.PUT("/team/project/info", handlers.SetProjectInfo)
	router.PUT("/team/project/invited", handlers.SetProjectInvited)
	router.GET("/team/project/info/invited", handlers.GetProjectInvitedInfo)
	router.POST("/team/project/exit", handlers.ExitProject)
	router.PUT("/team/project/member/perm", handlers.SetProjectMemberPermission)
	router.PUT("/team/project/creator", handlers.ChangeProjectCreator)
	router.DELETE("/team/project/member", handlers.RemoveProjectMember)
	router.PUT("/team/project/favorite", handlers.SetProjectFavorite)
	router.GET("/team/project/favorite/list", handlers.GetFavorProjectList)
	router.POST("/team/project/document/move", handlers.MoveDocument)
	// 反馈
	router.POST("/feedback", handlers.PostFeedback)
	// router.POST("/test/001", handlers.CreateTest)
	// router.GET("/storage_auth", handlers.CheckStorageAuth)
}
