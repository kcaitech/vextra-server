package document

import (
	"errors"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/common/response"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/providers/safereview"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils"
	"kcaitech.com/kcserver/utils/sliceutil"
	"kcaitech.com/kcserver/utils/str"
	myTime "kcaitech.com/kcserver/utils/time"
)

// CreateProject 创建项目
func CreateProject(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	projectService := services.NewProjectService()
	var req struct {
		TeamId      string `json:"team_id" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	teamId := (req.TeamId)
	if teamId == "" {
		response.BadRequest(c, "参数错误：team_id")
		return
	}
	if req.Name == "" {
		response.BadRequest(c, "参数错误：name")
		return
	}
	teamService := services.NewTeamService()
	var team models.Team
	if err := teamService.GetById(teamId, &team); err != nil {
		if errors.Is(err, services.ErrRecordNotFound) {
			response.BadRequest(c, "团队不存在")
		} else {
			response.ServerError(c, "查询错误")
		}
		return
	}
	// 权限校验
	teamMemberService := teamService.TeamMemberService
	var teamMember models.TeamMember
	if err := teamMemberService.Get(&teamMember, "team_id = ? and user_id = ?", teamId, userId); err != nil {
		response.Forbidden(c, "")
		return
	}
	if teamMember.PermType < models.TeamPermTypeEditable {
		response.Forbidden(c, "")
		return
	}
	reviewClient := services.GetSafereviewClient()
	if req.Name != "" && reviewClient != nil {
		reviewResponse, err := (reviewClient).ReviewText(req.Name)
		if err != nil || reviewResponse.Status != safereview.ReviewTextResultPass {
			log.Println("名称审核不通过", req.Name, err, reviewResponse)
			response.ReviewFail(c, "审核不通过")
			return
		}
	}
	if req.Description != "" && reviewClient != nil {
		reviewResponse, err := (reviewClient).ReviewText(req.Description)
		if err != nil || reviewResponse.Status != safereview.ReviewTextResultPass {
			log.Println("描述审核不通过", req.Description, err, reviewResponse)
			response.ReviewFail(c, "审核不通过")
			return
		}
	}

	id, err := utils.GenerateBase62ID()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	project := models.Project{
		TeamId:      teamId,
		Name:        req.Name,
		Description: req.Description,
		Id:          id,
	}
	if projectService.Create(&project) != nil {
		response.ServerError(c, "项目创建失败")
		return
	}

	projectMemberService := services.NewProjectMemberService()
	projectMember := models.ProjectMember{
		ProjectId:      project.Id,
		UserId:         userId,
		PermType:       models.ProjectPermTypeCreator,
		PermSourceType: models.ProjectPermSourceTypeCustom,
	}
	if projectMemberService.Create(&projectMember) != nil {
		response.ServerError(c, "项目创建失败.")
		return
	}

	response.Success(c, &project)
}

// GetProjectList 获取项目列表
func GetProjectList(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	teamId := (c.Query("team_id"))
	projectService := services.NewProjectService()
	result := projectService.FindProject(teamId, userId)
	// 获取用户信息
	type ProjectQueryWithCreator struct {
		Project             models.Project         `json:"project"`
		CreatorUser         models.UserProfile     `json:"creator"`
		CreatorTeamNickname string                 `json:"creator_team_nickname"`
		SelfPermType        models.ProjectPermType `json:"self_perm_type"`
		IsInTeam            bool                   `json:"is_in_team"`
		IsInvited           bool                   `json:"is_invited"`
	}

	userIds := make([]string, 0)
	for _, item := range result {
		userIds = append(userIds, item.CreatorTeamMember.UserId)
	}
	// 获取用户信息
	userMap, err := GetUsersInfo(c, userIds)
	if err != nil {
		log.Println("get users info fail:", err.Error())
		response.ServerError(c, "查询错误")
		return
	}
	resultWithCreator := make([]ProjectQueryWithCreator, 0)
	for _, item := range result {
		user, ok := userMap[item.CreatorTeamMember.UserId]
		if !ok {
			continue
		}
		resultWithCreator = append(resultWithCreator, ProjectQueryWithCreator{
			Project: item.Project,
			CreatorUser: models.UserProfile{
				Id:       user.UserID,
				Nickname: user.Profile.Nickname,
				Avatar:   user.Profile.Avatar,
			},
			CreatorTeamNickname: item.CreatorTeamNickname,
			SelfPermType:        item.SelfPermType,
			IsInTeam:            item.IsInTeam,
			IsInvited:           item.IsInvited,
		})
	}

	response.Success(c, resultWithCreator)
}

// GetProjectMemberList 获取项目成员列表
func GetProjectMemberList(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	projectId := (c.Query("project_id"))
	if projectId == "" {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	projectService := services.NewProjectService()
	if permType, err := projectService.GetProjectPermTypeByForUser(projectId, userId); err != nil || permType == nil {
		response.ServerError(c, "查询错误")
		return
	} else if *permType == models.ProjectPermTypeNone || *permType < models.ProjectPermTypeReadOnly {
		response.Forbidden(c, "")
		return
	}
	result := projectService.FindProjectMember(projectId)

	// 获取user信息
	userIds := make([]string, 0)
	for _, member := range result {
		userIds = append(userIds, member.ProjectMember.UserId)
	}

	userMap, err := GetUsersInfo(c, userIds)
	if err != nil {
		log.Println("get users info fail:", err.Error())
		response.ServerError(c, "查询错误")
		return
	}

	type User struct {
		Nickname string `json:"nickname"`
		Id       string `json:"id"`
		Avatar   string `json:"avatar"`
	}

	// 将用户信息与团队成员信息合并
	type MemberWithUser struct {
		Project       models.Project       `json:"project"`
		ProjectMember models.ProjectMember `json:"project_member"`
		User          User                 `json:"user"`
	}

	mergedResult := make([]MemberWithUser, 0)

	// 合并团队成员和用户信息
	for _, member := range result {
		userId := member.ProjectMember.UserId
		userInfo, exists := userMap[userId]

		if exists {
			user := userInfo
			mergedMember := MemberWithUser{
				Project:       member.Project,
				ProjectMember: member.ProjectMember,
				User: User{
					Id:       user.UserID,
					Nickname: user.Profile.Nickname,
					Avatar:   user.Profile.Avatar,
				},
			}
			mergedResult = append(mergedResult, mergedMember)
		}
	}

	response.Success(c, mergedResult)
}

// DeleteProject 删除项目
func DeleteProject(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	projectId := (c.Query("project_id"))
	if projectId == "" {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	projectService := services.NewProjectService()
	var project models.Project
	if err := projectService.GetById(projectId, &project); err != nil {
		if errors.Is(err, services.ErrRecordNotFound) {
			response.BadRequest(c, "项目不存在")
		} else {
			response.ServerError(c, "项目查询错误")
		}
		return
	}
	// 权限校验
	permType, err := projectService.GetProjectPermTypeByForUser(projectId, userId)
	if err != nil || permType == nil {
		response.ServerError(c, "权限查询错误")
		return
	}
	if *permType == models.ProjectPermTypeNone || *permType < models.ProjectPermTypeCreator {
		response.Forbidden(c, "")
		return
	}
	// 删除项目申请记录
	projectMemberService := projectService.ProjectMemberService
	if _, err := projectMemberService.Delete("project_id = ?", projectId); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.ServerError(c, "项目申请记录删除失败")
		return
	}
	// 删除项目成员
	if _, err := projectMemberService.Delete("project_id = ?", projectId); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.ServerError(c, "项目成员删除失败")
		return
	}
	// 删除项目
	if _, err := projectService.Delete("id = ?", projectId); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.ServerError(c, "项目删除失败")
		return
	}
	// 删除项目文档
	documentService := services.NewDocumentService()
	if _, err := documentService.Delete("project_id = ?", projectId); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.ServerError(c, "项目文档删除失败")
		return
	}
	response.Success(c, "")
}

// ApplyJoinProject 申请加入项目
func ApplyJoinProject(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		ProjectId      string `json:"project_id" binding:"required"`
		ApplicantNotes string `json:"applicant_notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}
	projectId := (req.ProjectId)
	if projectId == "" {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	projectService := services.NewProjectService()
	var project models.Project
	if err := projectService.GetById(projectId, &project); err != nil {
		if errors.Is(err, services.ErrRecordNotFound) {
			response.BadRequest(c, "项目不存在")
		} else {
			response.ServerError(c, "项目查询错误")
		}
		return
	}
	if !project.OpenInvite {
		response.BadRequest(c, "项目未开启邀请")
		return
	}
	permType, err := projectService.GetProjectPermTypeByForUser(projectId, userId)
	if err != nil || permType == nil {
		response.ServerError(c, "权限查询错误")
		return
	}
	if *permType != models.ProjectPermTypeNone {
		response.BadRequest(c, "已加入项目")
		return
	}
	projectJoinRequestService := projectService.ProjectJoinRequestService
	if ok, err := projectJoinRequestService.Exist("deleted_at is null and project_id = ? and user_id = ? and status = ?", projectId, userId, models.ProjectJoinRequestStatusPending); err != nil {
		response.ServerError(c, "申请查询错误")
		return
	} else if ok {
		response.BadRequest(c, "不能重复申请")
		return
	}
	var projectJoinRequest models.ProjectJoinRequest
	if project.NeedApproval {
		projectJoinRequest = models.ProjectJoinRequest{
			UserId:         userId,
			ProjectId:      projectId,
			PermType:       project.PermType,
			ApplicantNotes: req.ApplicantNotes,
		}
	} else {
		projectJoinRequest = models.ProjectJoinRequest{
			UserId:         userId,
			ProjectId:      projectId,
			PermType:       project.PermType,
			ApplicantNotes: req.ApplicantNotes,
			Status:         models.ProjectJoinRequestStatusApproved,
			ProcessedAt:    myTime.Time(time.Now()),
		}
	}
	if err := projectJoinRequestService.Create(&projectJoinRequest); err != nil {
		response.ServerError(c, "申请新建错误")
		return
	}
	if !project.NeedApproval {
		if err := projectService.ProjectMemberService.Create(&models.ProjectMember{
			UserId:         userId,
			ProjectId:      projectId,
			PermType:       project.PermType,
			PermSourceType: models.ProjectPermSourceTypeCustom,
		}); err != nil {
			response.ServerError(c, "权限新建错误")
			return
		}
	}
	response.Success(c, "")
}

// GetProjectJoinRequestList 获取申请列表
func GetProjectJoinRequestList(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	projectId := c.Query("project_id")
	// if projectId <= 0 {
	// 	projectId = 0
	// }
	startTimeStr := ""
	startTimeInt := str.DefaultToInt(c.Query("start_time"), 0)
	if startTimeInt > 0 {
		startTimeStr = myTime.Time(time.UnixMilli(startTimeInt)).String()
	}
	projectService := services.NewProjectService()
	projectJoinRequestMessageShowService := projectService.ProjectJoinRequestMessageShowService
	now := myTime.Time(time.Now())
	result := projectService.FindProjectJoinRequest(userId, projectId, startTimeStr)
	// 获取用户信息
	userIds := make([]string, 0)
	for _, item := range result {
		userIds = append(userIds, item.ProjectJoinRequest.UserId)
	}

	userMap, err := GetUsersInfo(c, userIds)
	if err != nil {
		response.ServerError(c, "获取用户信息失败")
		return
	}

	for i := range result {
		userId := result[i].ProjectJoinRequest.UserId
		userInfo, exists := userMap[userId]
		if exists {
			result[i].User = &models.UserProfile{
				Id:       userInfo.UserID,
				Nickname: userInfo.Profile.Nickname,
				Avatar:   userInfo.Profile.Avatar,
			}
		}
	}

	if startTimeStr == "" {
		response.Success(c, result)
		return
	}

	var messageShowList []models.ProjectJoinRequestMessageShow
	if err := projectJoinRequestMessageShowService.Find(&messageShowList, "user_id = ? and project_id = ?", userId, projectId); err != nil {
		log.Println("ProjectJoinRequestMessageShow查询错误", err)
		response.ServerError(c, "")
		return
	}
	result = sliceutil.FilterT(func(item services.ProjectJoinRequestQuery) bool {
		return sliceutil.Find(func(messageShowItem models.ProjectJoinRequestMessageShow) bool {
			return messageShowItem.ProjectJoinRequestId == item.ProjectJoinRequest.Id
		}, messageShowList...) == nil
	}, result...)
	newMessageShowList := sliceutil.MapT(func(item services.ProjectJoinRequestQuery) models.BaseModel {
		return &models.ProjectJoinRequestMessageShow{
			ProjectJoinRequestId: item.ProjectJoinRequest.Id,
			UserId:               userId,
			ProjectId:            projectId,
			FirstDisplayedAt:     now,
		}
	}, result...)
	for i := range newMessageShowList {
		if err := projectJoinRequestMessageShowService.Create(newMessageShowList[i]); err != nil {
			log.Println("ProjectJoinRequestMessageShow新建错误", err)
			response.ServerError(c, "")
			return
		}
	}
	response.Success(c, result)
}

// GetSelfProjectJoinRequestList 获取自身的项目加入申请列表
func GetSelfProjectJoinRequestList(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	projectId := c.Query("project_id")
	// if projectId <= 0 {
	// 	projectId = 0
	// }
	startTimeStr := ""
	startTimeInt := str.DefaultToInt(c.Query("start_time"), 0)
	if startTimeInt > 0 {
		startTimeStr = myTime.Time(time.UnixMilli(startTimeInt)).String()
	}
	projectService := services.NewProjectService()
	result := projectService.FindSelfProjectJoinRequest(userId, projectId, startTimeStr)

	// 获取用户信息
	userIds := make([]string, 0)
	for _, item := range result {
		if item.ProjectJoinRequest.ProcessedBy != "" {
			userIds = append(userIds, item.ProjectJoinRequest.ProcessedBy)
		}
	}

	if len(userIds) > 0 {
		userMap, err := GetUsersInfo(c, userIds)
		if err != nil {
			response.ServerError(c, "获取用户信息失败")
			return
		}

		for i := range result {
			if result[i].ProjectJoinRequest.ProcessedBy != "" {
				userInfo, exists := userMap[result[i].ProjectJoinRequest.ProcessedBy]
				if exists {
					result[i].User = &models.UserProfile{
						Id:       userInfo.UserID,
						Nickname: userInfo.Profile.Nickname,
						Avatar:   userInfo.Profile.Avatar,
					}
				}
			}
		}
	}

	response.Success(c, result)
}

// ReviewProjectJoinRequest 权限申请审核
func ReviewProjectJoinRequest(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		ApplyId      string `json:"apply_id" binding:"required"`
		ApprovalCode uint8  `json:"approval_code" binding:"min=0,max=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}
	projectJoinRequestsId := str.DefaultToInt(req.ApplyId, 0)
	if projectJoinRequestsId <= 0 {
		response.BadRequest(c, "参数错误：apply_id")
		return
	}
	approvalCode := req.ApprovalCode
	if approvalCode != 0 && approvalCode != 1 {
		response.BadRequest(c, "参数错误：approval_code")
		return
	}
	// 权限校验
	projectService := services.NewProjectService()
	var projectJoinRequest models.ProjectJoinRequest
	if err := projectService.ProjectJoinRequestService.Get(
		&projectJoinRequest,
		&services.JoinArgsRaw{
			Join: "inner join project_member on project_member.project_id = project_join_request.project_id" +
				" and project_member.user_id = ? and (project_member.perm_type = ? or project_member.perm_type = ?)" +
				" and project_member.deleted_at is null",
			Args: []any{userId, models.ProjectPermTypeAdmin, models.ProjectPermTypeCreator},
		},
		&services.WhereArgs{
			Query: "project_join_request.id = ? and project_join_request.status = ?",
			Args:  []interface{}{projectJoinRequestsId, models.ProjectJoinRequestStatusPending},
		},
	); err != nil {
		if errors.Is(err, services.ErrRecordNotFound) {
			response.BadRequest(c, "申请已被处理或无权限")
		} else {
			response.ServerError(c, "查询错误")
		}
		return
	}
	if projectJoinRequest.PermType < models.ProjectPermTypeReadOnly || projectJoinRequest.PermType > models.ProjectPermTypeEditable {
		response.BadRequest(c, "参数错误：projectJoinRequest.PermType")
		return
	}
	if approvalCode == 0 {
		projectJoinRequest.Status = models.ProjectJoinRequestStatusDenied
	} else if approvalCode == 1 {
		projectJoinRequest.Status = models.ProjectJoinRequestStatusApproved
	}
	projectJoinRequest.ProcessedAt = myTime.Time(time.Now())
	projectJoinRequest.ProcessedBy = userId
	if _, err := projectService.ProjectJoinRequestService.UpdatesById(projectJoinRequestsId, &projectJoinRequest); err != nil {
		response.ServerError(c, "更新错误")
		return
	}
	if approvalCode == 1 {
		if permType, err := projectService.GetProjectPermTypeByForUser(projectJoinRequest.ProjectId, projectJoinRequest.UserId); err != nil || permType == nil {
			response.ServerError(c, "查询错误")
			return
		} else if *permType == models.ProjectPermTypeNone {
			if err := projectService.ProjectMemberService.Create(&models.ProjectMember{
				ProjectId:      projectJoinRequest.ProjectId,
				UserId:         projectJoinRequest.UserId,
				PermType:       projectJoinRequest.PermType,
				PermSourceType: models.ProjectPermSourceTypeCustom,
			}); err != nil {
				response.ServerError(c, "新建错误")
				return
			}
		} else if projectJoinRequest.PermType <= *permType {
			response.Success(c, "")
			return
		} else {
			if _, err := projectService.ProjectMemberService.UpdatesIgnoreZero(&models.ProjectMember{
				PermType:       projectJoinRequest.PermType,
				PermSourceType: models.ProjectPermSourceTypeCustom,
			}, "project_id = ? and user_id = ?", projectJoinRequest.ProjectId, projectJoinRequest.UserId); err != nil {
				response.ServerError(c, "更新错误")
				return
			}
		}
	}
	response.Success(c, "")
}

// SetProjectInfo 设置项目信息
func SetProjectInfo(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		ProjectId   string `json:"project_id" binding:"required"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	projectId := (req.ProjectId)
	if projectId == "" {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	projectService := services.NewProjectService()
	if permType, err := projectService.GetProjectPermTypeByForUser(projectId, userId); err != nil || permType == nil {
		response.ServerError(c, "查询错误")
		return
	} else if *permType == models.ProjectPermTypeNone || *permType < models.ProjectPermTypeAdmin {
		response.Forbidden(c, "")
		return
	}
	if req.Name == "" && req.Description == "" {
		response.BadRequest(c, "")
		return
	}
	reviewClient := services.GetSafereviewClient()
	if req.Name != "" || req.Description != "" {
		if req.Name != "" && reviewClient != nil {
			reviewResponse, err := (reviewClient).ReviewText(req.Name)
			if err != nil || reviewResponse.Status != safereview.ReviewTextResultPass {
				log.Println("名称审核不通过", req.Name, err, reviewResponse)
				response.ReviewFail(c, "审核不通过")
				return
			}
		}
		if req.Description != "" && reviewClient != nil {
			reviewResponse, err := (reviewClient).ReviewText(req.Description)
			if err != nil || reviewResponse.Status != safereview.ReviewTextResultPass {
				log.Println("描述审核不通过", req.Description, err, reviewResponse)
				response.ReviewFail(c, "审核不通过")
				return
			}
		}
		if _, err := projectService.UpdatesIgnoreZeroById(projectId, &models.Project{
			Name:        req.Name,
			Description: req.Description,
		}); err != nil {
			response.ServerError(c, "更新错误")
			return
		}
	}
	response.Success(c, "")
}

// SetProjectInvited 修改项目邀请设置
func SetProjectInvited(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		ProjectId    string                  `json:"project_id" binding:"required"`
		IsPublic     *bool                   `json:"is_public"`     // 是否在团队内部公开
		PermType     *models.ProjectPermType `json:"perm_type"`     // 团队内的公开权限类型、或邀请权限类型
		OpenInvite   *bool                   `json:"open_invite"`   // 邀请开关
		NeedApproval *bool                   `json:"need_approval"` // 申请是否需要审批
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	projectId := (req.ProjectId)
	if projectId == "" {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	if req.IsPublic == nil && req.PermType == nil && req.OpenInvite == nil && req.NeedApproval == nil {
		response.BadRequest(c, "")
		return
	}
	if req.PermType != nil && (*req.PermType < models.ProjectPermTypeReadOnly || *req.PermType > models.ProjectPermTypeEditable) {
		response.BadRequest(c, "参数错误：perm_type")
		return
	}
	projectService := services.NewProjectService()
	if permType, err := projectService.GetProjectPermTypeByForUser(projectId, userId); err != nil || permType == nil {
		response.ServerError(c, "查询错误")
		return
	} else if *permType == models.ProjectPermTypeNone || *permType < models.ProjectPermTypeAdmin {
		response.Forbidden(c, "")
		return
	}
	updateColumns := map[string]any{}
	if req.IsPublic != nil {
		updateColumns["is_public"] = *req.IsPublic
	}
	if req.PermType != nil {
		updateColumns["perm_type"] = *req.PermType
	}
	if req.OpenInvite != nil {
		updateColumns["open_invite"] = *req.OpenInvite
	}
	if req.NeedApproval != nil {
		updateColumns["need_approval"] = *req.NeedApproval
	}
	if _, err := projectService.UpdateColumnsById(projectId, updateColumns); err != nil {
		response.ServerError(c, "更新错误")
		return
	}
	response.Success(c, "")
}

// GetProjectInvitedInfo 获取项目邀请信息
func GetProjectInvitedInfo(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	projectId := (c.Query("project_id"))
	if projectId == "" {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	projectService := services.NewProjectService()
	var project models.Project
	if err := projectService.GetById(projectId, &project); err != nil {
		log.Println("GetProjectInvitedInfo查询错误", err)
		response.ServerError(c, "查询错误")
		return
	}
	// if !project.OpenInvite {
	// 	response.Success(c, "项目邀请已关闭")
	// 	return
	// }
	selfPermType, err := projectService.GetProjectPermTypeByForUser(projectId, userId)
	if err != nil {
		log.Println("GetProjectInvitedInfo查询错误1", err)
		response.ServerError(c, "查询错误")
		return
	}
	result := map[string]any{
		"id":                project.Id,
		"name":              project.Name,
		"self_perm_type":    selfPermType,
		"invited_perm_type": project.PermType,
		"invited_switch":    project.OpenInvite,
	}
	response.Success(c, result)
}

// ExitProject 退出项目
func ExitProject(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		ProjectId string `json:"project_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}
	projectId := (req.ProjectId)
	if projectId == "" {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	projectService := services.NewProjectService()
	if permType, err := projectService.GetProjectPermTypeByForUser(projectId, userId); err != nil || permType == nil {
		response.ServerError(c, "查询错误")
		return
	} else if *permType == models.ProjectPermTypeNone || *permType == models.ProjectPermTypeCreator {
		response.Forbidden(c, "")
		return
	}
	// 删除项目成员
	projectMemberService := services.NewProjectMemberService()
	if _, err := projectMemberService.Delete("project_id = ? and user_id = ?", projectId, userId); err != nil {
		response.ServerError(c, "项目成员删除失败")
		return
	}
	response.Success(c, "")
}

// SetProjectMemberPermission 设置项目成员权限
func SetProjectMemberPermission(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		ProjectId string                  `json:"project_id" binding:"required"`
		UserId    string                  `json:"user_id" binding:"required"`
		PermType  *models.ProjectPermType `json:"perm_type" binding:"min=1,max=4"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}
	projectId := (req.ProjectId)
	if projectId == "" {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	reqUserId := req.UserId
	if reqUserId == "" {
		response.BadRequest(c, "参数错误：user_id")
		return
	}
	if req.PermType == nil {
		response.BadRequest(c, "参数错误：perm_type")
		return
	}
	projectService := services.NewProjectService()
	selfPermType, err := projectService.GetProjectPermTypeByForUser(projectId, userId)
	if err != nil {
		response.ServerError(c, "查询错误")
		return
	} else if selfPermType == nil || *selfPermType < models.ProjectPermTypeAdmin {
		response.Forbidden(c, "")
		return
	}
	permType, err := projectService.GetProjectPermTypeByForUser(projectId, reqUserId)
	if err != nil || permType == nil {
		response.ServerError(c, "查询错误")
		return
	}
	if *permType >= *selfPermType {
		response.Forbidden(c, "")
		return
	}
	projectMemberService := services.NewProjectMemberService()
	if _, err := projectMemberService.UpdateColumns(map[string]any{
		"perm_type": req.PermType,
	}, "project_id = ? and user_id = ?", projectId, reqUserId); err != nil {
		response.ServerError(c, "更新错误")
		return
	}
	response.Success(c, "")
}

// ChangeProjectCreator 更改项目创建者
func ChangeProjectCreator(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		ProjectId string `json:"project_id" binding:"required"`
		UserId    string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}
	projectId := (req.ProjectId)
	if projectId == "" {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	reqUserId := req.UserId
	if reqUserId == "" {
		response.BadRequest(c, "参数错误：user_id")
		return
	}
	if userId == reqUserId {
		response.BadRequest(c, "不能转移给自己")
		return
	}
	projectService := services.NewProjectService()
	selfPermType, err := projectService.GetProjectPermTypeByForUser(projectId, userId)
	if err != nil {
		response.ServerError(c, "查询错误")
		return
	} else if selfPermType == nil || *selfPermType != models.ProjectPermTypeCreator {
		response.Forbidden(c, "")
		return
	}
	projectMemberService := services.NewProjectMemberService()
	transactDB := projectMemberService.DBModule.DB.Begin() // 开启事务
	needRollback := false
	defer func() {
		if needRollback {
			transactDB.Rollback()
		} else {
			transactDB.Commit()
		}
	}()
	if err := transactDB.Error; err != nil {
		response.ServerError(c, "更新错误")
		needRollback = true
		return
	}
	if _, err := projectMemberService.UpdateColumns(map[string]any{
		"perm_type": models.ProjectPermTypeAdmin,
	}, "project_id = ? and user_id = ?", projectId, userId); err != nil {
		response.ServerError(c, "更新错误.")
		needRollback = true
		return
	}
	if _, err := projectMemberService.UpdateColumns(map[string]any{
		"perm_type": models.ProjectPermTypeCreator,
	}, "project_id = ? and user_id = ?", projectId, reqUserId); err != nil {
		response.ServerError(c, "更新错误..")
		needRollback = true
		return
	}
	response.Success(c, "")
}

// RemoveProjectMember 移除项目成员
func RemoveProjectMember(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	projectId := (c.Query("project_id"))
	if projectId == "" {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	reqUserId := c.Query("user_id")
	if reqUserId == "" {
		response.BadRequest(c, "参数错误：user_id")
		return
	}
	if userId == reqUserId {
		response.BadRequest(c, "不能移除自己")
		return
	}
	projectService := services.NewProjectService()
	selfPermType, err := projectService.GetProjectPermTypeByForUser(projectId, userId)
	if err != nil {
		response.ServerError(c, "查询错误")
		return
	} else if selfPermType == nil || *selfPermType < models.ProjectPermTypeAdmin {
		response.Forbidden(c, "")
		return
	}
	permType, err := projectService.GetProjectPermTypeByForUser(projectId, reqUserId)
	if err != nil || permType == nil {
		response.ServerError(c, "查询错误")
		return
	} else if *permType >= *selfPermType {
		response.Forbidden(c, "")
		return
	}
	// 删除项目成员
	projectMemberService := services.NewProjectMemberService()
	if _, err := projectMemberService.Delete("project_id = ? and user_id = ?", projectId, reqUserId); err != nil {
		response.ServerError(c, "团队成员删除失败")
		return
	}
	response.Success(c, "")
}

// SetProjectFavorite 收藏/取消收藏项目
func SetProjectFavorite(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		ProjectId string `json:"project_id" binding:"required"`
		IsFavor   *bool  `json:"is_favor" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}
	isFavor := *req.IsFavor
	projectId := (req.ProjectId)
	if projectId == "" {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	projectService := services.NewProjectService()
	if err := projectService.ToggleProjectFavorite(userId, projectId, isFavor); err != nil {
		response.ServerError(c, "操作失败")
		return
	}
	response.Success(c, "")
}

// GetFavorProjectList 获取收藏项目列表
func GetFavorProjectList(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	teamId := (c.Query("team_id"))
	projectService := services.NewProjectService()
	result := projectService.FindFavorProject(teamId, userId)
	response.Success(c, result)
}

// MoveDocument 移动文档
func MoveDocument(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		SourceProjectId string `json:"source_project_id"`
		TargetProjectId string `json:"target_project_id"`
		DocumentId      string `json:"document_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}

	documentId := (req.DocumentId)
	if documentId == "" {
		response.BadRequest(c, "参数错误：document_id")
		return
	}
	var document models.Document
	documentService := services.NewDocumentService()
	if err := documentService.GetById(documentId, &document); err != nil {
		response.ServerError(c, "查询错误")
		return
	}

	sourceProjectId := (req.SourceProjectId)
	targetProjectId := (req.TargetProjectId)
	if (sourceProjectId == "" && targetProjectId == "") || sourceProjectId == targetProjectId {
		response.BadRequest(c, "参数错误：source_project_id、target_project_id")
		return
	}
	if document.ProjectId != sourceProjectId {
		response.BadRequest(c, "参数错误：document.project_id")
		return
	}
	if document.ProjectId == "" && document.UserId != userId {
		response.Forbidden(c, "")
		return
	}

	projectService := services.NewProjectService()
	if sourceProjectId != "" {
		sourcePermType, err := projectService.GetProjectPermTypeByForUser(sourceProjectId, userId)
		if err != nil {
			response.ServerError(c, "查询错误")
			return
		}
		if sourcePermType == nil || *sourcePermType < models.ProjectPermTypeEditable {
			response.Forbidden(c, "")
			return
		}
	}

	var targetTeamId string
	if targetProjectId != "" {
		targetPermType, err := projectService.GetProjectPermTypeByForUser(targetProjectId, userId)
		if err != nil {
			response.ServerError(c, "查询错误")
			return
		}
		if targetPermType == nil || *targetPermType < models.ProjectPermTypeEditable {
			response.Forbidden(c, "")
			return
		}
		var targetProject models.Project
		if err := projectService.GetById(targetProjectId, &targetProject); err != nil {
			response.ServerError(c, "查询错误")
			return
		}
		targetTeamId = targetProject.TeamId
	}

	if _, err := documentService.UpdateColumns(map[string]any{
		"team_id":    targetTeamId,
		"project_id": targetProjectId,
	}, "id = ?", documentId); err != nil {
		response.ServerError(c, "更新错误")
		return
	}

	response.Success(c, "")
}
