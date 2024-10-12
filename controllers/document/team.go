package document

import (
	"encoding/base64"
	"errors"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/common/gin/auth"
	"kcaitech.com/kcserver/common/gin/response"
	"kcaitech.com/kcserver/common/models"
	"kcaitech.com/kcserver/common/safereview"
	safereviewBase "kcaitech.com/kcserver/common/safereview/base"
	"kcaitech.com/kcserver/common/services"
	config "kcaitech.com/kcserver/controllers"
	"kcaitech.com/kcserver/utils/sliceutil"
	"kcaitech.com/kcserver/utils/str"
	myTime "kcaitech.com/kcserver/utils/time"
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

	reviewResponse, err := safereview.Client.ReviewText(req.Name)
	if err != nil || reviewResponse.Status != safereviewBase.ReviewTextResultPass {
		log.Println("名称审核不通过", req.Name, err, reviewResponse)
		response.Fail(c, "审核不通过")
		return
	}
	if req.Description != "" {
		reviewResponse, err = safereview.Client.ReviewText(req.Description)
		if err != nil || reviewResponse.Status != safereviewBase.ReviewTextResultPass {
			log.Println("描述审核不通过", req.Description, err, reviewResponse)
			response.Fail(c, "审核不通过")
			return
		}
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
		if fileHeader.Size > 2<<20 {
			response.BadRequest(c, "文件大小不能超过2MB")
			return
		}
		file, err := fileHeader.Open()
		if err != nil {
			response.BadRequest(c, "获取文件失败")
			return
		}
		defer file.Close()
		fileBytes := make([]byte, fileHeader.Size)
		if _, err := file.Read(fileBytes); err != nil {
			response.BadRequest(c, "读取文件失败")
			return
		}
		contentType := fileHeader.Header.Get("Content-Type")
		base64Str := base64.StdEncoding.EncodeToString(fileBytes)
		reviewResponse, err := safereview.Client.ReviewPictureFromBase64(base64Str)
		if err != nil {
			log.Println("头像审核错误", err)
			response.Fail(c, "头像审核错误")
			return
		} else if reviewResponse.Status != safereviewBase.ReviewImageResultPass {
			log.Println("头像审核不通过", err, reviewResponse)
			response.Fail(c, "头像审核不通过")
			return
		}
		_, _ = services.NewTeamService().UploadTeamAvatar(&team, fileBytes, contentType)
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
		"id":          str.IntToString(team.Id),
		"name":        team.Name,
		"description": team.Description,
	}
	if team.Avatar != "" {
		result["avatar"] = config.Config.StorageUrl.Attatch + team.Avatar
	}
	response.Success(c, result)
}

// GetTeamList 获取团队列表
func GetTeamList(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	teamService := services.NewTeamService()
	result := teamService.FindTeamByUserId(userId)
	response.Success(c, result)
}

// GetTeamMemberList 获取团队成员列表
func GetTeamMemberList(c *gin.Context) {
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
	if permType, err := teamService.GetTeamPermTypeByForUser(teamId, userId); err != nil || permType == nil {
		response.Fail(c, "查询错误")
		return
	} else if *permType == models.TeamPermTypeNone || *permType < models.TeamPermTypeReadOnly {
		response.Forbidden(c, "")
		return
	}
	result := teamService.FindTeamMember(teamId)
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
	if _, err := teamJoinRequestService.Delete("team_id = ?", teamId); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.Fail(c, "团队申请记录删除失败")
		return
	}
	// 删除团队成员
	if _, err := teamMemberService.Delete("team_id = ?", teamId); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.Fail(c, "团队成员删除失败")
		return
	}
	// 删除团队项目
	projectService := services.NewProjectService()
	if _, err := projectService.Delete("team_id = ?", teamId); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.Fail(c, "团队项目删除失败")
		return
	}
	// 删除团队文档
	documentService := services.NewDocumentService()
	if _, err := documentService.Delete("team_id = ?", teamId); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		response.Fail(c, "团队文档删除失败")
		return
	}
	// 删除团队
	if _, err := teamService.Delete("id = ?", teamId); err != nil {
		response.Fail(c, "团队删除失败")
		return
	}
	response.Success(c, "")
}

// ApplyJoinTeam 申请加入团队
func ApplyJoinTeam(c *gin.Context) {
	type FailCode int
	const (
		FailCodeTeamNotExist FailCode = iota + 1
		FailCodeInvitedNotOpen
		FailCodeAlreadyJoined
		FailCodeAlreadyApplied
	)
	failResponseData := map[string]any{}

	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		TeamId         string `json:"team_id" binding:"required"`
		ApplicantNotes string `json:"applicant_notes"`
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
			failResponseData["code"] = FailCodeTeamNotExist
			response.BadRequestData(c, "团队不存在", failResponseData)
		} else {
			response.Fail(c, "查询错误")
		}
		return
	}
	if !team.InvitedSwitch {
		failResponseData["code"] = FailCodeInvitedNotOpen
		response.BadRequestData(c, "团队未开启邀请", failResponseData)
		return
	}
	invitedPermType := team.InvitedPermType
	teamMemberService := teamService.TeamMemberService
	if ok, err := teamMemberService.Exist("deleted_at is null and team_id = ? and user_id = ?", teamId, userId); ok {
		failResponseData["code"] = FailCodeAlreadyJoined
		response.BadRequestData(c, "已加入团队", failResponseData)
		return
	} else if err != nil {
		response.Fail(c, "查询错误.")
		return
	}
	teamJoinRequestService := teamService.TeamJoinRequestService
	if ok, err := teamJoinRequestService.Exist("deleted_at is null and team_id = ? and user_id = ? and status = ?", teamId, userId, models.TeamJoinRequestStatusPending); ok {
		failResponseData["code"] = FailCodeAlreadyApplied
		response.BadRequestData(c, "不能重复申请", failResponseData)
		return
	} else if err != nil {
		response.Fail(c, "查询错误..")
		return
	}
	if err := teamJoinRequestService.Create(&models.TeamJoinRequest{
		UserId:         userId,
		TeamId:         teamId,
		PermType:       invitedPermType,
		ApplicantNotes: req.ApplicantNotes,
	}); err != nil {
		response.Fail(c, "新建错误")
	}
	response.Success(c, "")
}

// GetTeamJoinRequestList 获取申请列表
func GetTeamJoinRequestList(c *gin.Context) {
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
	teamJoinRequestMessageShowService := teamService.TeamJoinRequestMessageShowService
	now := myTime.Time(time.Now())
	result := teamService.FindTeamJoinRequest(userId, teamId, startTimeStr)
	if startTimeStr == "" {
		response.Success(c, result)
		return
	}
	var messageShowList []models.TeamJoinRequestMessageShow
	if err := teamJoinRequestMessageShowService.Find(&messageShowList, "user_id = ? and team_id = ?", userId, teamId); err != nil {
		log.Println("TeamJoinRequestMessageShow查询错误", err)
		response.Fail(c, "")
		return
	}
	result = sliceutil.FilterT(func(item services.TeamJoinRequestQuery) bool {
		return sliceutil.Find(func(messageShowItem models.TeamJoinRequestMessageShow) bool {
			return messageShowItem.TeamJoinRequestId == item.TeamJoinRequest.Id
		}, messageShowList...) == nil
	}, result...)
	newMessageShowList := sliceutil.MapT(func(item services.TeamJoinRequestQuery) models.ModelData {
		return &models.TeamJoinRequestMessageShow{
			TeamJoinRequestId: item.TeamJoinRequest.Id,
			UserId:            userId,
			TeamId:            teamId,
			FirstDisplayedAt:  now,
		}
	}, result...)
	for i := range newMessageShowList {
		if err := teamJoinRequestMessageShowService.Create(newMessageShowList[i]); err != nil {
			log.Println("TeamJoinRequestMessageShow新建错误", err)
			response.Fail(c, "")
			return
		}
	}
	response.Success(c, result)
}

// GetSelfTeamJoinRequestList 获取自身的申请列表
func GetSelfTeamJoinRequestList(c *gin.Context) {
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
	result := teamService.FindSelfTeamJoinRequest(userId, teamId, startTimeStr)
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
			Join: "inner join team_member on team_member.team_id = team_join_request.team_id" +
				" and team_member.user_id = ? and (team_member.perm_type = ? or team_member.perm_type = ?)" +
				" and team_member.deleted_at is null",
			Args: []any{userId, models.TeamPermTypeAdmin, models.TeamPermTypeCreator},
		},
		&services.WhereArgs{
			Query: "team_join_request.id = ? and team_join_request.status = ?",
			Args:  []interface{}{teamJoinRequestsId, models.TeamJoinRequestStatusPending},
		},
	); err != nil {
		if errors.Is(err, services.ErrRecordNotFound) {
			response.BadRequest(c, "申请已被处理或无权限")
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
	if _, err := teamService.TeamJoinRequestService.UpdatesById(teamJoinRequestsId, &teamJoinRequest); err != nil {
		response.Fail(c, "更新错误")
		return
	}
	if approvalCode == 1 {
		if permType, err := teamService.GetTeamPermTypeByForUser(teamJoinRequest.TeamId, teamJoinRequest.UserId); err != nil || permType == nil {
			response.Fail(c, "查询错误")
			return
		} else if *permType == models.TeamPermTypeNone {
			if err := teamService.TeamMemberService.Create(&models.TeamMember{
				TeamId:   teamJoinRequest.TeamId,
				UserId:   teamJoinRequest.UserId,
				PermType: teamJoinRequest.PermType,
			}); err != nil {
				response.Fail(c, "新建错误")
				return
			}
		} else if teamJoinRequest.PermType <= *permType {
			response.Success(c, "")
			return
		} else {
			if _, err := teamService.TeamMemberService.UpdatesIgnoreZero(&models.TeamMember{
				PermType: teamJoinRequest.PermType,
			}, "team_id = ? and user_id = ?", teamJoinRequest.TeamId, teamJoinRequest.UserId); err != nil {
				response.Fail(c, "更新错误")
				return
			}
		}
	}
	response.Success(c, "")
}

// SetTeamInfo 设置团队信息
func SetTeamInfo(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		TeamId      string `json:"team_id" form:"team_id" binding:"required"`
		Name        string `json:"name" form:"name"`
		Description string `json:"description" form:"description"`
	}
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	teamId := str.DefaultToInt(req.TeamId, 0)
	if teamId <= 0 {
		response.BadRequest(c, "参数错误：team_id")
		return
	}
	teamService := services.NewTeamService()
	if permType, err := teamService.GetTeamPermTypeByForUser(teamId, userId); err != nil || permType == nil {
		response.Fail(c, "查询错误")
		return
	} else if *permType < models.TeamPermTypeAdmin {
		response.Forbidden(c, "")
		return
	}
	fileHeader, err := c.FormFile("avatar")
	fileExists := err == nil
	if fileExists && fileHeader.Size > 2<<20 {
		response.BadRequest(c, "文件大小不能超过2MB")
		return
	}
	if req.Name == "" && req.Description == "" && !fileExists {
		response.BadRequest(c, "")
		return
	}
	if req.Name != "" || req.Description != "" {
		if req.Name != "" {
			reviewResponse, err := safereview.Client.ReviewText(req.Name)
			if err != nil || reviewResponse.Status != safereviewBase.ReviewTextResultPass {
				log.Println("名称审核不通过", req.Name, err, reviewResponse)
				response.Fail(c, "审核不通过")
				return
			}
		}
		if req.Description != "" {
			reviewResponse, err := safereview.Client.ReviewText(req.Description)
			if err != nil || reviewResponse.Status != safereviewBase.ReviewTextResultPass {
				log.Println("描述审核不通过", req.Description, err, reviewResponse)
				response.Fail(c, "审核不通过")
				return
			}
		}
		if _, err := teamService.UpdatesIgnoreZeroById(teamId, &models.Team{
			Name:        req.Name,
			Description: req.Description,
		}); err != nil {
			response.Fail(c, "更新错误")
			return
		}
	}
	result := map[string]any{}
	if fileExists {
		file, err := fileHeader.Open()
		if err != nil {
			response.BadRequest(c, "获取文件失败")
			return
		}
		defer file.Close()
		fileBytes := make([]byte, fileHeader.Size)
		if _, err := file.Read(fileBytes); err != nil {
			response.BadRequest(c, "读取文件失败")
			return
		}
		contentType := fileHeader.Header.Get("Content-Type")
		base64Str := base64.StdEncoding.EncodeToString(fileBytes)
		reviewResponse, err := safereview.Client.ReviewPictureFromBase64(base64Str)
		if err != nil {
			log.Println("头像审核错误", err)
			response.Fail(c, "头像审核错误")
			return
		} else if reviewResponse.Status != safereviewBase.ReviewImageResultPass {
			log.Println("头像审核不通过", err, reviewResponse)
			response.Fail(c, "头像审核不通过")
			return
		}
		if avatarPath, err := teamService.UploadTeamAvatarById(teamId, fileBytes, contentType); err == nil {
			result["avatar"] = config.Config.StorageUrl.Attatch + avatarPath
		}
	}
	response.Success(c, result)
}

// SetTeamInvited 修改团队邀请设置
func SetTeamInvited(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		TeamId          string               `json:"team_id" binding:"required"`
		InvitedPermType *models.TeamPermType `json:"invited_perm_type"`
		InvitedSwitch   *bool                `json:"invited_switch"`
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
	if req.InvitedPermType == nil && req.InvitedSwitch == nil {
		response.BadRequest(c, "")
		return
	}
	if req.InvitedPermType != nil && (*req.InvitedPermType < models.TeamPermTypeReadOnly || *req.InvitedPermType > models.TeamPermTypeEditable) {
		response.BadRequest(c, "参数错误：invited_perm_type")
		return
	}
	teamService := services.NewTeamService()
	if permType, err := teamService.GetTeamPermTypeByForUser(teamId, userId); err != nil || permType == nil {
		response.Fail(c, "查询错误")
		return
	} else if *permType < models.TeamPermTypeAdmin {
		response.Forbidden(c, "")
		return
	}
	updateColumns := map[string]any{}
	if req.InvitedPermType != nil {
		updateColumns["invited_perm_type"] = *req.InvitedPermType
	}
	if req.InvitedSwitch != nil {
		updateColumns["invited_switch"] = *req.InvitedSwitch
	}
	if _, err := teamService.UpdateColumnsById(teamId, updateColumns); err != nil {
		response.Fail(c, "更新错误")
		return
	}
	response.Success(c, "")
}

// GetTeamInvitedInfo 获取团队信息
func GetTeamInvitedInfo(c *gin.Context) {
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
		response.Fail(c, "查询错误")
		return
	}
	selfPermType, err := teamService.GetTeamPermTypeByForUser(teamId, userId)
	if err != nil {
		response.Fail(c, "查询错误")
		return
	}
	result := map[string]any{
		"id":                str.IntToString(team.Id),
		"name":              team.Name,
		"self_perm_type":    selfPermType,
		"invited_perm_type": team.InvitedPermType,
		"invited_switch":    team.InvitedSwitch,
	}
	response.Success(c, result)
}

// ExitTeam 退出团队
func ExitTeam(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		TeamId string `json:"team_id" binding:"required"`
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
	if permType, err := teamService.GetTeamPermTypeByForUser(teamId, userId); err != nil || permType == nil {
		response.Fail(c, "查询错误")
		return
	} else if *permType == models.TeamPermTypeCreator {
		response.Forbidden(c, "")
		return
	}
	// 退出或删除项目 todo

	// 删除团队成员
	teamMemberService := services.NewTeamMemberService()
	if _, err := teamMemberService.Delete("team_id = ? and user_id = ?", teamId, userId); err != nil {
		response.Fail(c, "团队成员删除失败")
		return
	}
	response.Success(c, "")
}

// SetTeamMemberPermission 设置团队成员权限
func SetTeamMemberPermission(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		TeamId   string               `json:"team_id" binding:"required"`
		UserId   string               `json:"user_id" binding:"required"`
		PermType *models.TeamPermType `json:"perm_type" binding:"min=0,max=2"`
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
	reqUserId := str.DefaultToInt(req.UserId, 0)
	if reqUserId <= 0 {
		response.BadRequest(c, "参数错误：user_id")
		return
	}
	if req.PermType == nil {
		response.BadRequest(c, "参数错误：perm_type")
		return
	}
	teamService := services.NewTeamService()
	selfPermType, err := teamService.GetTeamPermTypeByForUser(teamId, userId)
	if err != nil {
		response.Fail(c, "查询错误")
		return
	} else if selfPermType == nil || *selfPermType < models.TeamPermTypeAdmin {
		response.Forbidden(c, "")
		return
	}
	permType, err := teamService.GetTeamPermTypeByForUser(teamId, reqUserId)
	if err != nil || permType == nil {
		response.Fail(c, "查询错误")
		return
	}
	if *permType >= *selfPermType {
		response.Forbidden(c, "")
		return
	}
	teamMemberService := services.NewTeamMemberService()
	if _, err := teamMemberService.UpdateColumns(map[string]any{
		"perm_type": req.PermType,
	}, "team_id = ? and user_id = ?", teamId, reqUserId); err != nil {
		response.Fail(c, "更新错误")
		return
	}
	response.Success(c, "")
}

// ChangeTeamCreator 更改团队创建者
func ChangeTeamCreator(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		TeamId string `json:"team_id" binding:"required"`
		UserId string `json:"user_id" binding:"required"`
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
	reqUserId := str.DefaultToInt(req.UserId, 0)
	if reqUserId <= 0 {
		response.BadRequest(c, "参数错误：user_id")
		return
	}
	if userId == reqUserId {
		response.BadRequest(c, "不能转移给自己")
		return
	}
	teamService := services.NewTeamService()
	selfPermType, err := teamService.GetTeamPermTypeByForUser(teamId, userId)
	if err != nil {
		response.Fail(c, "查询错误")
		return
	} else if selfPermType == nil || *selfPermType != models.TeamPermTypeCreator {
		response.Forbidden(c, "")
		return
	}
	teamMemberService := services.NewTeamMemberService()
	teamMemberService.DB = teamMemberService.DB.Begin() // 开启事务
	needRollback := false
	defer func() {
		if needRollback {
			teamMemberService.DB.Rollback()
		} else {
			teamMemberService.DB.Commit()
		}
	}()
	if err := teamMemberService.DB.Error; err != nil {
		response.Fail(c, "更新错误")
		needRollback = true
		return
	}
	if _, err := teamMemberService.UpdateColumns(map[string]any{
		"perm_type": models.TeamPermTypeAdmin,
	}, "team_id = ? and user_id = ?", teamId, userId); err != nil {
		response.Fail(c, "更新错误.")
		needRollback = true
		return
	}
	if _, err := teamMemberService.UpdateColumns(map[string]any{
		"perm_type": models.TeamPermTypeCreator,
	}, "team_id = ? and user_id = ?", teamId, reqUserId); err != nil {
		response.Fail(c, "更新错误..")
		needRollback = true
		return
	}
	response.Success(c, "")
}

// RemoveTeamMember 移除团队成员
func RemoveTeamMember(c *gin.Context) {
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
	reqUserId := str.DefaultToInt(c.Query("user_id"), 0)
	if reqUserId <= 0 {
		response.BadRequest(c, "参数错误：user_id")
		return
	}
	if userId == reqUserId {
		response.BadRequest(c, "不能移除自己")
		return
	}
	teamService := services.NewTeamService()
	selfPermType, err := teamService.GetTeamPermTypeByForUser(teamId, userId)
	if err != nil {
		response.Fail(c, "查询错误")
		return
	} else if selfPermType == nil || *selfPermType < models.TeamPermTypeAdmin {
		response.Forbidden(c, "")
		return
	}
	permType, err := teamService.GetTeamPermTypeByForUser(teamId, reqUserId)
	if err != nil || permType == nil {
		response.Fail(c, "查询错误")
		return
	} else if *permType >= *selfPermType {
		response.Forbidden(c, "")
		return
	}
	// 退出或删除项目 todo

	// 删除团队成员
	teamMemberService := services.NewTeamMemberService()
	if _, err := teamMemberService.Delete("team_id = ? and user_id = ?", teamId, reqUserId); err != nil {
		response.Fail(c, "团队成员删除失败")
		return
	}
	response.Success(c, "")
}

// SetTeamMemberNickname 设置团队成员昵称
func SetTeamMemberNickname(c *gin.Context) {
	userId, err := auth.GetUserId(c)
	if err != nil {
		response.Unauthorized(c)
		return
	}
	var req struct {
		TeamId   string `json:"team_id" binding:"required"`
		UserId   string `json:"user_id" binding:"required"`
		Nickname string `json:"nickname" binding:"required"`
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
	reqUserId := str.DefaultToInt(req.UserId, 0)
	if reqUserId <= 0 {
		response.BadRequest(c, "参数错误：user_id")
		return
	}
	teamService := services.NewTeamService()
	teamMemberService := teamService.TeamMemberService
	if ok, err := teamMemberService.Exist("team_id = ? and user_id = ?", teamId, reqUserId); err != nil {
		response.Fail(c, "查询错误")
		return
	} else if !ok {
		response.BadRequest(c, "团队成员不存在")
		return
	}
	if userId != reqUserId {
		selfPermType, err := teamService.GetTeamPermTypeByForUser(teamId, userId)
		if err != nil {
			response.Fail(c, "查询错误")
			return
		} else if selfPermType == nil || *selfPermType < models.TeamPermTypeAdmin {
			response.Forbidden(c, "")
			return
		}
	}

	reviewResponse, err := safereview.Client.ReviewText(req.Nickname)
	if err != nil || reviewResponse.Status != safereviewBase.ReviewTextResultPass {
		log.Println("昵称审核不通过", req.Nickname, err, reviewResponse)
		response.Fail(c, "审核不通过")
		return
	}

	if _, err := teamMemberService.UpdateColumns(map[string]any{
		"nickname": req.Nickname,
	}, "team_id = ? and user_id = ?", teamId, reqUserId); err != nil {
		response.Fail(c, "更新错误")
		return
	}
	response.Success(c, "")
}
