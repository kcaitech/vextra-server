package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/utils/str"
	myTime "protodesign.cn/kcserver/utils/time"
	"time"
)

// CreateProject 创建项目
func CreateProject(c *gin.Context) {
	userId, err := auth.GetUserId(c)
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
	teamId := str.DefaultToInt(req.TeamId, 0)
	if teamId <= 0 {
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
			response.Fail(c, "查询错误")
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
	project := models.Project{
		TeamId:      teamId,
		Name:        req.Name,
		Description: req.Description,
	}
	if projectService.Create(&project) != nil {
		response.Fail(c, "项目创建失败")
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
		response.Fail(c, "项目创建失败.")
		return
	}

	response.Success(c, &project)
}

// GetProjectList 获取项目列表
func GetProjectList(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	teamId := str.DefaultToInt(c.Query("team_id"), 0)
	projectService := services.NewProjectService()
	result := projectService.FindProject(teamId, userId)
	response.Success(c, result)
}

// GetProjectMemberList 获取项目成员列表
func GetProjectMemberList(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	projectId := str.DefaultToInt(c.Query("project_id"), 0)
	if projectId <= 0 {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	projectService := services.NewProjectService()
	if permType, err := projectService.GetProjectPermTypeByForUser(projectId, userId); err != nil {
		response.Fail(c, "查询错误")
		return
	} else if permType == nil || *permType < models.ProjectPermTypeReadOnly {
		response.Forbidden(c, "")
		return
	}
	result := projectService.FindProjectMember(projectId)
	response.Success(c, result)
}

// DeleteProject 删除项目
func DeleteProject(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	projectId := str.DefaultToInt(c.Query("project_id"), 0)
	if projectId <= 0 {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	projectService := services.NewProjectService()
	var project models.Project
	if err := projectService.GetById(projectId, &project); err != nil {
		if errors.Is(err, services.ErrRecordNotFound) {
			response.BadRequest(c, "项目不存在")
		} else {
			response.Fail(c, "项目查询错误")
		}
		return
	}
	// 权限校验
	permType, err := projectService.GetProjectPermTypeByForUser(projectId, userId)
	if err != nil {
		response.Fail(c, "权限查询错误")
		return
	}
	if permType == nil || *permType < models.ProjectPermTypeAdmin {
		response.Forbidden(c, "")
		return
	}
	// 删除项目申请记录
	projectMemberService := projectService.ProjectMemberService
	if _, err := projectMemberService.Delete("project_id = ?", projectId); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.Fail(c, "项目申请记录删除失败")
		return
	}
	// 删除项目成员
	if _, err := projectMemberService.Delete("project_id = ?", projectId); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.Fail(c, "项目成员删除失败")
		return
	}
	// 删除项目
	if _, err := projectService.Delete("id = ?", projectId); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.Fail(c, "项目删除失败")
		return
	}
	// 删除项目文档
	documentService := services.NewDocumentService()
	if _, err := documentService.Delete("project_id = ?", projectId); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.Fail(c, "项目文档删除失败")
		return
	}
	response.Success(c, "")
}

// ApplyJoinProject 申请加入项目
func ApplyJoinProject(c *gin.Context) {
	userId, err := auth.GetUserId(c)
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
	projectId := str.DefaultToInt(req.ProjectId, 0)
	if projectId <= 0 {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	projectService := services.NewProjectService()
	var project models.Project
	if err := projectService.GetById(projectId, &project); err != nil {
		if errors.Is(err, services.ErrRecordNotFound) {
			response.BadRequest(c, "项目不存在")
		} else {
			response.Fail(c, "项目查询错误")
		}
		return
	}
	if !project.InvitedSwitch {
		response.BadRequest(c, "项目未开启邀请")
		return
	}
	permType, err := projectService.GetProjectPermTypeByForUser(projectId, userId)
	if err != nil {
		response.Fail(c, "权限查询错误")
		return
	}
	if permType != nil {
		response.BadRequest(c, "已加入项目")
		return
	}
	projectJoinRequestService := projectService.ProjectJoinRequestService
	if ok, err := projectJoinRequestService.Exist("deleted_at is null and project_id = ? and user_id = ? and status = ?", projectId, userId, models.ProjectJoinRequestStatusPending); err != nil {
		response.Fail(c, "申请查询错误")
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
		response.Fail(c, "申请新建错误")
		return
	}
	if !project.NeedApproval {
		if err := projectService.ProjectMemberService.Create(&models.ProjectMember{
			UserId:         userId,
			ProjectId:      projectId,
			PermType:       project.PermType,
			PermSourceType: models.ProjectPermSourceTypeCustom,
		}); err != nil {
			response.Fail(c, "权限新建错误")
			return
		}
	}
	response.Success(c, "")
}

// GetProjectJoinRequestList 获取申请列表
func GetProjectJoinRequestList(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	projectId := str.DefaultToInt(c.Query("project_id"), 0)
	if projectId <= 0 {
		projectId = 0
	}
	startTimeStr := ""
	startTimeInt := str.DefaultToInt(c.Query("start_time"), 0)
	if startTimeInt > 0 {
		startTimeStr = myTime.Time(time.UnixMilli(startTimeInt)).String()
	}
	projectService := services.NewProjectService()
	result := projectService.FindProjectJoinRequest(userId, projectId, startTimeStr)
	response.Success(c, result)
}

// ReviewProjectJoinRequest 权限申请审核
func ReviewProjectJoinRequest(c *gin.Context) {
	userId, err := auth.GetUserId(c)
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
			response.Fail(c, "查询错误")
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
		response.Fail(c, "更新错误")
		return
	}
	if approvalCode == 1 {
		if permType, err := projectService.GetProjectPermTypeByForUser(projectJoinRequest.ProjectId, projectJoinRequest.UserId); err != nil {
			response.Fail(c, "查询错误")
			return
		} else if permType == nil {
			if err := projectService.ProjectMemberService.Create(&models.ProjectMember{
				ProjectId:      projectJoinRequest.ProjectId,
				UserId:         projectJoinRequest.UserId,
				PermType:       projectJoinRequest.PermType,
				PermSourceType: models.ProjectPermSourceTypeCustom,
			}); err != nil {
				response.Fail(c, "新建错误")
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
				response.Fail(c, "更新错误")
				return
			}
		}
	}
	response.Success(c, "")
}

// SetProjectInfo 设置项目信息
func SetProjectInfo(c *gin.Context) {
	userId, err := auth.GetUserId(c)
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
	projectId := str.DefaultToInt(req.ProjectId, 0)
	if projectId <= 0 {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	projectService := services.NewProjectService()
	if permType, err := projectService.GetProjectPermTypeByForUser(projectId, userId); err != nil {
		response.Fail(c, "查询错误")
		return
	} else if permType == nil || *permType < models.ProjectPermTypeAdmin {
		response.Forbidden(c, "")
		return
	}
	if req.Name == "" && req.Description == "" {
		response.BadRequest(c, "")
		return
	}
	if req.Name != "" || req.Description != "" {
		if _, err := projectService.UpdatesIgnoreZeroById(projectId, &models.Project{
			Name:        req.Name,
			Description: req.Description,
		}); err != nil {
			response.Fail(c, "更新错误")
			return
		}
	}
	response.Success(c, "")
}

// SetProjectInvited 修改项目邀请设置
func SetProjectInvited(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		ProjectId     string                  `json:"project_id" binding:"required"`
		PublicSwitch  *bool                   `json:"public_switch"`  // 是否在团队内部公开
		PermType      *models.ProjectPermType `json:"perm_type"`      // 团队内的公开权限类型、或邀请权限类型
		InvitedSwitch *bool                   `json:"invited_switch"` // 邀请开关
		NeedApproval  *bool                   `json:"need_approval"`  // 申请是否需要审批
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	projectId := str.DefaultToInt(req.ProjectId, 0)
	if projectId <= 0 {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	if req.PublicSwitch == nil && req.PermType == nil && req.InvitedSwitch == nil && req.NeedApproval == nil {
		response.BadRequest(c, "")
		return
	}
	if req.PermType != nil && (*req.PermType < models.ProjectPermTypeReadOnly || *req.PermType > models.ProjectPermTypeEditable) {
		response.BadRequest(c, "参数错误：perm_type")
		return
	}
	projectService := services.NewProjectService()
	if permType, err := projectService.GetProjectPermTypeByForUser(projectId, userId); err != nil {
		response.Fail(c, "查询错误")
		return
	} else if permType == nil || *permType < models.ProjectPermTypeAdmin {
		response.Forbidden(c, "")
		return
	}
	updateColumns := map[string]any{}
	if req.PublicSwitch != nil {
		updateColumns["public_switch"] = *req.PublicSwitch
	}
	if req.PermType != nil {
		updateColumns["perm_type"] = *req.PermType
	}
	if req.InvitedSwitch != nil {
		updateColumns["invited_switch"] = *req.InvitedSwitch
	}
	if req.NeedApproval != nil {
		updateColumns["need_approval"] = *req.NeedApproval
	}
	if _, err := projectService.UpdateColumnsById(projectId, updateColumns); err != nil {
		response.Fail(c, "更新错误")
		return
	}
	response.Success(c, "")
}

// GetProjectInvitedInfo 获取项目邀请信息
func GetProjectInvitedInfo(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	projectId := str.DefaultToInt(c.Query("project_id"), 0)
	if projectId <= 0 {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	projectService := services.NewProjectService()
	var project models.Project
	if err := projectService.GetById(projectId, &project); err != nil {
		response.Fail(c, "查询错误")
		return
	}
	if !project.InvitedSwitch {
		response.Fail(c, "项目邀请已关闭")
		return
	}
	selfPermType, err := projectService.GetProjectPermTypeByForUser(projectId, userId)
	if err != nil {
		response.Fail(c, "查询错误")
		return
	}
	result := map[string]any{
		"id":                project.Id,
		"name":              project.Name,
		"self_perm_type":    selfPermType,
		"invited_perm_type": project.PermType,
	}
	response.Success(c, result)
}

// ExitProject 退出项目
func ExitProject(c *gin.Context) {
	userId, err := auth.GetUserId(c)
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
	projectId := str.DefaultToInt(req.ProjectId, 0)
	if projectId <= 0 {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	projectService := services.NewProjectService()
	if permType, err := projectService.GetProjectPermTypeByForUser(projectId, userId); err != nil {
		response.Fail(c, "查询错误")
		return
	} else if permType == nil || *permType == models.ProjectPermTypeCreator {
		response.Forbidden(c, "")
		return
	}
	// 删除项目成员
	projectMemberService := services.NewProjectMemberService()
	if _, err := projectMemberService.Delete("project_id = ? and user_id = ?", projectId, userId); err != nil {
		response.Fail(c, "项目成员删除失败")
		return
	}
	response.Success(c, "")
}

// SetProjectMemberPermission 设置项目成员权限
func SetProjectMemberPermission(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		ProjectId string                  `json:"project_id" binding:"required"`
		UserId    string                  `json:"user_id" binding:"required"`
		PermType  *models.ProjectPermType `json:"perm_type" binding:"min=1,max=3"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}
	projectId := str.DefaultToInt(req.ProjectId, 0)
	if projectId <= 0 {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	reqUserId := str.DefaultToInt(req.UserId, 0)
	if reqUserId <= 0 {
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
		response.Fail(c, "查询错误")
		return
	} else if selfPermType == nil || *selfPermType < models.ProjectPermTypeAdmin {
		response.Forbidden(c, "")
		return
	}
	permType, err := projectService.GetProjectPermTypeByForUser(projectId, reqUserId)
	if err != nil {
		response.Fail(c, "查询错误")
		return
	}
	if permType != nil && *permType >= *selfPermType {
		response.Forbidden(c, "")
		return
	}
	projectMemberService := services.NewProjectMemberService()
	if _, err := projectMemberService.UpdateColumns(map[string]any{
		"perm_type": req.PermType,
	}, "project_id = ? and user_id = ?", projectId, reqUserId); err != nil {
		response.Fail(c, "更新错误")
		return
	}
	response.Success(c, "")
}

// ChangeProjectCreator 更改项目创建者
func ChangeProjectCreator(c *gin.Context) {
	userId, err := auth.GetUserId(c)
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
	projectId := str.DefaultToInt(req.ProjectId, 0)
	if projectId <= 0 {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	reqUserId := str.DefaultToInt(req.UserId, 0)
	if reqUserId <= 0 {
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
		response.Fail(c, "查询错误")
		return
	} else if selfPermType == nil || *selfPermType != models.ProjectPermTypeCreator {
		response.Forbidden(c, "")
		return
	}
	projectMemberService := services.NewProjectMemberService()
	projectMemberService.DB = projectMemberService.DB.Begin() // 开启事务
	needRollback := false
	defer func() {
		if needRollback {
			projectMemberService.DB.Rollback()
		} else {
			projectMemberService.DB.Commit()
		}
	}()
	if err := projectMemberService.DB.Error; err != nil {
		response.Fail(c, "更新错误")
		needRollback = true
		return
	}
	if _, err := projectMemberService.UpdateColumns(map[string]any{
		"perm_type": models.ProjectPermTypeAdmin,
	}, "project_id = ? and user_id = ?", projectId, userId); err != nil {
		response.Fail(c, "更新错误.")
		needRollback = true
		return
	}
	if _, err := projectMemberService.UpdateColumns(map[string]any{
		"perm_type": models.ProjectPermTypeCreator,
	}, "project_id = ? and user_id = ?", projectId, reqUserId); err != nil {
		response.Fail(c, "更新错误..")
		needRollback = true
		return
	}
	response.Success(c, "")
}

// RemoveProjectMember 移除项目成员
func RemoveProjectMember(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	projectId := str.DefaultToInt(c.Query("project_id"), 0)
	if projectId <= 0 {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	reqUserId := str.DefaultToInt(c.Query("user_id"), 0)
	if reqUserId <= 0 {
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
		response.Fail(c, "查询错误")
		return
	} else if selfPermType == nil || *selfPermType < models.ProjectPermTypeAdmin {
		response.Forbidden(c, "")
		return
	}
	permType, err := projectService.GetProjectPermTypeByForUser(projectId, reqUserId)
	if err != nil {
		response.Fail(c, "查询错误")
		return
	} else if permType != nil && *permType >= *selfPermType {
		response.Forbidden(c, "")
		return
	}
	// 删除项目成员
	projectMemberService := services.NewProjectMemberService()
	if _, err := projectMemberService.Delete("project_id = ? and user_id = ?", projectId, reqUserId); err != nil {
		response.Fail(c, "团队成员删除失败")
		return
	}
	response.Success(c, "")
}

// SetProjectFavorite 收藏/取消收藏项目
func SetProjectFavorite(c *gin.Context) {
	userId, err := auth.GetUserId(c)
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
	projectId := str.DefaultToInt(req.ProjectId, 0)
	if projectId <= 0 {
		response.BadRequest(c, "参数错误：project_id")
		return
	}
	projectService := services.NewProjectService()
	if err := projectService.ToggleProjectFavorite(userId, projectId, isFavor); err != nil {
		response.Fail(c, "操作失败")
		return
	}
	response.Success(c, "")
}

// GetFavorProjectList 获取收藏项目列表
func GetFavorProjectList(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	teamId := str.DefaultToInt(c.Query("team_id"), 0)
	projectService := services.NewProjectService()
	result := projectService.FindFavorProject(teamId, userId)
	response.Success(c, result)
}

// MoveDocument 移动文档
func MoveDocument(c *gin.Context) {
	userId, err := auth.GetUserId(c)
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

	documentId := str.DefaultToInt(req.DocumentId, 0)
	if documentId <= 0 {
		response.BadRequest(c, "参数错误：document_id")
		return
	}
	var document models.Document
	documentService := services.NewDocumentService()
	if err := documentService.GetById(documentId, &document); err != nil {
		response.Fail(c, "查询错误")
		return
	}

	sourceProjectId := str.DefaultToInt(req.SourceProjectId, 0)
	targetProjectId := str.DefaultToInt(req.TargetProjectId, 0)
	if (sourceProjectId <= 0 && targetProjectId <= 0) || sourceProjectId == targetProjectId {
		response.BadRequest(c, "参数错误：source_project_id、target_project_id")
		return
	}
	if document.ProjectId != sourceProjectId {
		response.BadRequest(c, "参数错误：document.project_id")
		return
	}
	if document.ProjectId == 0 && document.UserId != userId {
		response.Forbidden(c, "")
		return
	}

	projectService := services.NewProjectService()
	if sourceProjectId != 0 {
		sourcePermType, err := projectService.GetProjectPermTypeByForUser(sourceProjectId, userId)
		if err != nil {
			response.Fail(c, "查询错误")
			return
		}
		if sourcePermType == nil || *sourcePermType < models.ProjectPermTypeEditable {
			response.Forbidden(c, "")
			return
		}
	}

	var targetTeamId int64
	if targetProjectId != 0 {
		targetPermType, err := projectService.GetProjectPermTypeByForUser(targetProjectId, userId)
		if err != nil {
			response.Fail(c, "查询错误")
			return
		}
		if targetPermType == nil || *targetPermType < models.ProjectPermTypeEditable {
			response.Forbidden(c, "")
			return
		}
		var targetProject models.Project
		if err := projectService.GetById(targetProjectId, &targetProject); err != nil {
			response.Fail(c, "查询错误")
			return
		}
		targetTeamId = targetProject.TeamId
	}

	if _, err := documentService.UpdateColumns(map[string]any{
		"team_id":    targetTeamId,
		"project_id": targetProjectId,
	}, "id = ?", documentId); err != nil {
		response.Fail(c, "更新错误")
		return
	}

	response.Success(c, "")
}
