package http

import (
	"github.com/gin-gonic/gin"
	handlers "kcaitech.com/kcserver/handlers/document"
)

func loadTeamRoutes(api *gin.RouterGroup) {
	router := api.Group("/team")
	// team
	router.POST("/", handlers.CreateTeam)                          // 创建团队
	router.GET("/list", handlers.GetTeamList)                      // 获取团队列表
	router.GET("/member/list", handlers.GetTeamMemberList)         // 获取团队成员列表
	router.DELETE("/", handlers.DeleteTeam)                        // 删除团队
	router.POST("/apply", handlers.ApplyJoinTeam)                  // 申请加入团队
	router.GET("/apply", handlers.GetTeamJoinRequestList)          // 获取团队申请列表
	router.GET("/self_apply", handlers.GetSelfTeamJoinRequestList) // 获取团队的申请通知信息
	router.POST("/apply/audit", handlers.ReviewTeamJoinRequest)    // 团队加入审核
	router.PUT("/info", handlers.SetTeamInfo)                      // 设置团队信息
	router.PUT("/invite", handlers.SetTeamInvited)                 // 设置团队邀请选项
	router.GET("/info/invite", handlers.GetTeamInvitedInfo)        // 获取团队信息
	router.POST("/exit", handlers.ExitTeam)                        // 退出团队
	router.PUT("/member/perm", handlers.SetTeamMemberPermission)   // 设置团队成员权限
	router.PUT("/owner", handlers.ChangeTeamCreator)               // 转移团队创建者
	router.DELETE("/member", handlers.RemoveTeamMember)            // 删除团队成员
	router.PUT("/member/nickname", handlers.SetTeamMemberNickname) // 设置团队成员昵称

	loadProjectRoutes(router)
}

func loadProjectRoutes(api *gin.RouterGroup) {
	router := api.Group("/project")
	// team/project
	router.POST("/", handlers.CreateProject)                          // 创建项目
	router.GET("/list", handlers.GetProjectList)                      // 获取项目列表
	router.GET("/member/list", handlers.GetProjectMemberList)         // 获取项目成员列表
	router.DELETE("/", handlers.DeleteProject)                        // 删除项目
	router.POST("/apply", handlers.ApplyJoinProject)                  // 申请加入项目
	router.GET("/apply", handlers.GetProjectJoinRequestList)          // 获取项目申请列表
	router.GET("/self_apply", handlers.GetSelfProjectJoinRequestList) // 获取项目的申请通知信息
	router.POST("/apply/audit", handlers.ReviewProjectJoinRequest)    // 项目加入审核
	router.PUT("/info", handlers.SetProjectInfo)                      // 设置项目信息
	router.PUT("/invite", handlers.SetProjectInvited)                 // 设置项目邀请信息
	router.GET("/info/invite", handlers.GetProjectInvitedInfo)        // 获取项目邀请信息
	router.POST("/exit", handlers.ExitProject)                        // 退出项目
	router.PUT("/member/perm", handlers.SetProjectMemberPermission)   // 设置项目成员权限
	router.PUT("/owner", handlers.ChangeProjectCreator)               // 转让项目创建者
	router.DELETE("/member", handlers.RemoveProjectMember)            // 将成员移出项目组
	router.PUT("/favorite", handlers.SetProjectFavorite)              // 是否收藏
	router.GET("/favorite/list", handlers.GetFavorProjectList)        // 获取项目收藏列表
	router.POST("/document/move", handlers.MoveDocument)              // 移动文档
}
