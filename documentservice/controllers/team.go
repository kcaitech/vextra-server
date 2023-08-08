package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"protodesign.cn/kcserver/common"
	"protodesign.cn/kcserver/common/gin/auth"
	"protodesign.cn/kcserver/common/gin/response"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/services"
	"protodesign.cn/kcserver/utils/str"
	myTime "protodesign.cn/kcserver/utils/time"
	"time"
)

// CreateTeam 创建团队
func CreateTeam(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	teamService := services.NewTeamService()
	var req struct {
		Name        string `json:"name" form:"name" binding:"required"`
		Description string `json:"description" form:"description"`
	}
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.Name == "" {
		response.BadRequest(c, "参数错误：name")
		return
	}
	team := models.Team{
		Name:        req.Name,
		Description: req.Description,
		Uid:         str.GetUid(),
	}
	if teamService.Create(&team) != nil {
		response.Fail(c, "团队创建失败")
		return
	}

	fileHeader, err := c.FormFile("avatar")
	if err == nil {
		file, err := fileHeader.Open()
		if err != nil {
			response.BadRequest(c, "获取文件失败")
			return
		}
		defer file.Close()
		fileSize := fileHeader.Size
		contentType := fileHeader.Header.Get("Content-Type")
		_, _ = services.NewTeamService().UploadTeamAvatar(&team, file, fileSize, contentType)
	}

	teamMemberService := services.NewTeamMemberService()
	teamMember := models.TeamMember{
		TeamId:   team.Id,
		UserId:   userId,
		PermType: models.TeamPermTypeCreator,
	}
	if teamMemberService.Create(&teamMember) != nil {
		response.Fail(c, "团队创建失败.")
		return
	}

	result := map[string]any{
		"id":          team.Id,
		"name":        team.Name,
		"description": team.Description,
	}
	if team.Avatar != "" {
		result["avatar"] = common.StorageHost + team.Avatar
	}
	response.Success(c, result)
}

// DeleteTeam 解散团队
func DeleteTeam(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	teamId := str.DefaultToInt(c.Query("team_id"), 0)
	if teamId <= 0 {
		response.BadRequest(c, "参数错误：team_id")
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
	if teamMember.PermType != models.TeamPermTypeCreator {
		response.Forbidden(c, "")
		return
	}
	// 删除团队申请记录
	teamJoinRequestService := teamService.TeamJoinRequestService
	if err := teamJoinRequestService.Delete("team_id = ?", teamId); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.Fail(c, "团队申请记录删除失败")
		return
	}
	// 删除团队成员
	if err := teamMemberService.Delete("team_id = ?", teamId); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.Fail(c, "团队成员删除失败")
		return
	}
	// 删除团队项目
	projectService := services.NewProjectService()
	if err := projectService.Delete("team_id = ?", teamId); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.Fail(c, "团队项目删除失败")
		return
	}
	// 删除团队文档
	documentService := services.NewDocumentService()
	if err := documentService.Delete("team_id = ?", teamId); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.Fail(c, "团队文档删除失败")
		return
	}
	// 删除团队
	if err := teamService.Delete("id = ?", teamId); err != nil {
		response.Fail(c, "团队删除失败")
		return
	}
	response.Success(c, "")
}

// ApplyJoinTeam 申请加入团队
func ApplyJoinTeam(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		TeamId         string              `json:"team_id" binding:"required"`
		PermType       models.TeamPermType `json:"perm_type" binding:"min=0,max=1"`
		ApplicantNotes string              `json:"applicant_notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "")
		return
	}
	teamId := str.DefaultToInt(req.TeamId, 0)
	if teamId <= 0 {
		response.BadRequest(c, "参数错误：team_id")
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
	teamMemberService := teamService.TeamMemberService
	if ok, err := teamMemberService.Exist("team_id = ? AND user_id = ?", teamId, userId); ok {
		response.BadRequest(c, "已加入团队")
		return
	} else if err != nil {
		response.Fail(c, "查询错误.")
		return
	}
	teamJoinRequestService := teamService.TeamJoinRequestService
	if ok, err := teamJoinRequestService.Exist("team_id = ? AND user_id = ?", teamId, userId); ok {
		response.BadRequest(c, "不能重复申请")
		return
	} else if err != nil {
		response.Fail(c, "查询错误..")
		return
	}
	permType := req.PermType
	if permType < models.TeamPermTypeReadOnly || permType > models.TeamPermTypeEditable {
		response.BadRequest(c, "参数错误：perm_type")
		return
	}
	if err := teamJoinRequestService.Create(&models.TeamJoinRequest{
		UserId:         userId,
		TeamId:         teamId,
		PermType:       permType,
		ApplicantNotes: req.ApplicantNotes,
	}); err != nil {
		response.Fail(c, "新建错误")
	}
	response.Success(c, "")
}

// GetTeamJoinRequestsList 获取申请列表
func GetTeamJoinRequestsList(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	teamId := str.DefaultToInt(c.Query("team_id"), 0)
	if teamId <= 0 {
		teamId = 0
	}
	startTimeStr := ""
	startTimeInt := str.DefaultToInt(c.Query("start_time"), 0)
	if startTimeInt > 0 {
		startTimeStr = myTime.Time(time.UnixMilli(startTimeInt)).String()
	}
	teamService := services.NewTeamService()
	result := teamService.FindTeamJoinRequests(userId, teamId, startTimeStr)
	response.Success(c, result)
}

// ReviewTeamJoinRequest 权限申请审核
func ReviewTeamJoinRequest(c *gin.Context) {
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
	teamJoinRequestsId := str.DefaultToInt(req.ApplyId, 0)
	if teamJoinRequestsId <= 0 {
		response.BadRequest(c, "参数错误：apply_id")
		return
	}
	approvalCode := req.ApprovalCode
	if approvalCode != 0 && approvalCode != 1 {
		response.BadRequest(c, "参数错误：approval_code")
		return
	}
	// 权限校验
	teamService := services.NewTeamService()
	var teamJoinRequest models.TeamJoinRequest
	if err := teamService.TeamJoinRequestService.Get(
		&teamJoinRequest,
		&services.JoinArgsRaw{
			Join: "inner join team_member on team_member.team_id = team_join_request.team_id and team_member.user_id = ? and (team_member.perm_type = ? or team_member.perm_type = ?)",
			Args: []any{userId, models.TeamPermTypeAdmin, models.TeamPermTypeCreator},
		},
		&services.WhereArgs{
			Query: "team_join_request.id = ? and team_join_request.status = ?",
			Args:  []interface{}{teamJoinRequestsId, models.StatusTypePending},
		},
	); err != nil {
		if errors.Is(err, services.ErrRecordNotFound) {
			response.BadRequest(c, "申请已被处理")
		} else {
			response.Fail(c, "查询错误")
		}
		return
	}
	if teamJoinRequest.PermType < models.TeamPermTypeReadOnly || teamJoinRequest.PermType > models.TeamPermTypeEditable {
		response.BadRequest(c, "参数错误：teamJoinRequest.PermType")
		return
	}
	if approvalCode == 0 {
		teamJoinRequest.Status = models.TeamJoinRequestStatusDenied
	} else if approvalCode == 1 {
		teamJoinRequest.Status = models.TeamJoinRequestStatusApproved
	}
	teamJoinRequest.ProcessedAt = myTime.Time(time.Now())
	teamJoinRequest.ProcessedBy = userId
	if err := teamService.TeamJoinRequestService.UpdatesById(teamJoinRequestsId, &teamJoinRequest); err != nil {
		response.Fail(c, "更新错误")
		return
	}
	if approvalCode == 1 {
		var teamMember models.TeamMember
		err := teamService.TeamMemberService.Get(&teamMember, "team_id = ? AND user_id = ?", teamJoinRequest.TeamId, teamJoinRequest.UserId)
		if err != nil && !errors.Is(err, services.ErrRecordNotFound) {
			response.Fail(c, "查询错误")
			return
		}
		if errors.Is(err, services.ErrRecordNotFound) {
			if err := teamService.TeamMemberService.Create(&models.TeamMember{
				TeamId:   teamJoinRequest.TeamId,
				UserId:   teamJoinRequest.UserId,
				PermType: teamJoinRequest.PermType,
			}); err != nil {
				response.Fail(c, "新建错误")
				return
			}
		} else {
			if teamJoinRequest.PermType <= teamMember.PermType {
				response.Success(c, "")
				return
			} else {
				teamMember.PermType = teamJoinRequest.PermType
				if err := teamService.TeamMemberService.UpdatesById(teamMember.Id, &teamMember); err != nil {
					response.Fail(c, "更新错误")
					return
				}
			}
		}
	}
	response.Success(c, "")
}
