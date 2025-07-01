package document

import (
	"encoding/base64"
	"errors"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"kcaitech.com/kcserver/handlers/common"
	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/providers/safereview"
	"kcaitech.com/kcserver/services"
	"kcaitech.com/kcserver/utils"
	"kcaitech.com/kcserver/utils/sliceutil"
	"kcaitech.com/kcserver/utils/str"
	myTime "kcaitech.com/kcserver/utils/time"
)

// CreateTeam 创建团队
func CreateTeam(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	teamService := services.NewTeamService()
	var req struct {
		Name        string `json:"name" form:"name" binding:"required"`
		Description string `json:"description" form:"description"`
	}
	if err := c.ShouldBind(&req); err != nil {
		common.BadRequest(c, err.Error())
		return
	}
	if req.Name == "" {
		common.BadRequest(c, "参数错误：name")
		return
	}

	reviewClient := services.GetSafereviewClient()
	if reviewClient != nil {
		reviewResponse, err := (reviewClient).ReviewText(req.Name)
		if err != nil {
			log.Println("名称审核失败", req.Name, err)
			common.ReviewFail(c, "审核失败")
			return
		} else if reviewResponse != nil && reviewResponse.Status != safereview.ReviewTextResultPass {
			log.Println("名称审核不通过", req.Name, reviewResponse)
			common.ReviewFail(c, "审核不通过")
			return
		}
		if req.Description != "" {
			reviewResponse, err = (reviewClient).ReviewText(req.Description)
			if err != nil {
				log.Println("描述审核失败", req.Description, err)
				common.ReviewFail(c, "审核失败")
				return
			} else if reviewResponse != nil && reviewResponse.Status != safereview.ReviewTextResultPass {
				log.Println("描述审核不通过", req.Description, reviewResponse)
				common.ReviewFail(c, "审核不通过")
				return
			}
		}
	}
	id, err := utils.GenerateBase62ID()
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}
	team := models.Team{
		Name:        req.Name,
		Description: req.Description,
		Id:          id,
	}
	if teamService.Create(&team) != nil {
		common.ServerError(c, "团队创建失败")
		return
	}

	fileHeader, err := c.FormFile("avatar")
	if err == nil {
		if fileHeader.Size > 2<<20 {
			common.BadRequest(c, "文件大小不能超过2MB")
			return
		}
		file, err := fileHeader.Open()
		if err != nil {
			common.BadRequest(c, "获取文件失败")
			return
		}
		defer file.Close()
		fileBytes := make([]byte, fileHeader.Size)
		if _, err := file.Read(fileBytes); err != nil {
			common.BadRequest(c, "读取文件失败")
			return
		}
		contentType := fileHeader.Header.Get("Content-Type")
		base64Str := base64.StdEncoding.EncodeToString(fileBytes)
		if reviewClient != nil {
			reviewResponse, err := (reviewClient).ReviewPictureFromBase64(base64Str)
			if err != nil {
				log.Println("头像审核错误", err)
				common.BadRequest(c, "头像审核错误")
				return
			} else if reviewResponse.Status != safereview.ReviewImageResultPass {
				log.Println("头像审核不通过", err, reviewResponse)
				common.BadRequest(c, "头像审核不通过")
				return
			}
		}
		_, err = services.NewTeamService().UploadTeamAvatar(&team, fileBytes, contentType)
		if err != nil {
			log.Println("头像上传失败", err)
			common.BadRequest(c, "头像上传失败")
			return
		}
	}

	teamMemberService := services.NewTeamMemberService()
	teamMember := models.TeamMember{
		TeamId:   team.Id,
		UserId:   userId,
		PermType: models.TeamPermTypeCreator,
	}
	if teamMemberService.Create(&teamMember) != nil {
		common.ServerError(c, "团队创建失败.")
		return
	}

	result := map[string]any{
		"id":          (team.Id),
		"name":        team.Name,
		"description": team.Description,
	}
	// todo
	// if team.Avatar != "" {
	// 	result["avatar"] = config.Config.StorageUrl.Attatch + team.Avatar
	// }
	common.Success(c, result)
}

// GetTeamList 获取团队列表
func GetTeamList(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	teamService := services.NewTeamService()
	result := teamService.FindTeamByUserId(userId)

	userIds := make([]string, 0)
	for _, item := range result {
		userIds = append(userIds, item.CreatorTeamMember.UserId)
	}
	// 获取用户信息
	userMap, err := GetUsersInfo(c, userIds)
	if err != nil {
		log.Println("get users info fail:", err.Error())
		common.ServerError(c, "查询错误")
		return
	}

	type TeamQueryResItemWithCreator struct {
		Team         models.Team         `json:"team"`
		CreatorUser  models.UserProfile  `json:"creator"`
		SelfPermType models.TeamPermType `json:"self_perm_type"`
	}

	resultWithCreator := make([]TeamQueryResItemWithCreator, 0)
	for _, item := range result {
		user, ok := userMap[item.CreatorTeamMember.UserId]
		if !ok {
			continue
		}
		resultWithCreator = append(resultWithCreator, TeamQueryResItemWithCreator{
			Team: models.Team{
				Id:              item.Team.Id,
				Name:            item.Team.Name,
				Description:     item.Team.Description,
				Avatar:          services.GetConfig().StorageUrl.Attatch + item.Team.Avatar,
				InvitedPermType: item.Team.InvitedPermType,
				OpenInvite:      item.Team.OpenInvite,
			},
			SelfPermType: item.SelfPermType,
			CreatorUser: models.UserProfile{
				Id:       user.UserID,
				Nickname: user.Nickname,
				Avatar:   user.Avatar,
			},
		})
	}

	common.Success(c, resultWithCreator)
}

// GetTeamMemberList 获取团队成员列表
func GetTeamMemberList(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	teamId := (c.Query("team_id"))
	if teamId == "" {
		common.BadRequest(c, "参数错误：team_id")
		return
	}
	teamService := services.NewTeamService()
	if permType, err := teamService.GetTeamPermTypeByForUser(teamId, userId); err != nil || permType == nil {
		common.ServerError(c, "查询错误")
		return
	} else if *permType == models.TeamPermTypeNone || *permType < models.TeamPermTypeReadOnly {
		common.Forbidden(c, "")
		return
	}
	result := teamService.FindTeamMember(teamId)
	// 获取user信息
	userIds := make([]string, 0)
	for _, member := range result {
		userIds = append(userIds, member.TeamMember.UserId)
	}
	userMap, err := GetUsersInfo(c, userIds)
	if err != nil {
		log.Println("get users info fail:", err.Error())
		common.ServerError(c, "查询错误")
		return
	}

	type User struct {
		Nickname string `json:"nickname"`
		Id       string `json:"id"`
		Avatar   string `json:"avatar"`
	}

	// 将用户信息与团队成员信息合并
	type MemberWithUser struct {
		Team       models.Team       `json:"team"`
		TeamMember models.TeamMember `json:"team_member"`
		User       User              `json:"user"`
	}

	mergedResult := make([]MemberWithUser, 0)

	// 合并团队成员和用户信息
	for _, member := range result {
		userId := member.TeamMember.UserId
		userInfo, exists := userMap[userId]

		if exists {
			user := userInfo
			mergedMember := MemberWithUser{
				Team:       member.Team,
				TeamMember: member.TeamMember,
				User: User{
					Id:       user.UserID,
					Nickname: user.Nickname,
					Avatar:   user.Avatar,
				},
			}
			mergedResult = append(mergedResult, mergedMember)
		}
	}

	common.Success(c, mergedResult)
}

// DeleteTeam 解散团队
func DeleteTeam(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	teamId := (c.Query("team_id"))
	if teamId == "" {
		common.BadRequest(c, "参数错误：team_id")
		return
	}
	teamService := services.NewTeamService()
	var team models.Team
	if err := teamService.GetById(teamId, &team); err != nil {
		if errors.Is(err, services.ErrRecordNotFound) {
			common.BadRequest(c, "团队不存在")
		} else {
			common.ServerError(c, "查询错误")
		}
		return
	}
	// 权限校验
	teamMemberService := teamService.TeamMemberService
	var teamMember models.TeamMember
	if err := teamMemberService.Get(&teamMember, "team_id = ? and user_id = ?", teamId, userId); err != nil {
		if errors.Is(err, services.ErrRecordNotFound) {
			// 记录不存在
			log.Println("团队成员不存在", err, teamId, userId)
			common.Forbidden(c, "不是团队成员")
			return
		} else {
			common.ServerError(c, "查询错误")
			return
		}
	}
	if teamMember.PermType != models.TeamPermTypeCreator {
		common.Forbidden(c, "团队所有者才可删除团队")
		return
	}
	// 删除团队申请记录
	teamJoinRequestService := teamService.TeamJoinRequestService
	if _, err := teamJoinRequestService.Delete("team_id = ?", teamId); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		common.ServerError(c, "团队申请记录删除失败")
		return
	}
	// 删除团队成员
	if _, err := teamMemberService.Delete("team_id = ?", teamId); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		common.ServerError(c, "团队成员删除失败")
		return
	}
	// 删除团队项目
	projectService := services.NewProjectService()
	if _, err := projectService.Delete("team_id = ?", teamId); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		common.ServerError(c, "团队项目删除失败")
		return
	}
	// 删除团队文档
	documentService := services.NewDocumentService()
	if _, err := documentService.Delete("team_id = ?", teamId); err != nil && !errors.Is(err, services.ErrRecordNotFound) {
		log.Println("团队文档删除失败", err)
		common.ServerError(c, "团队文档删除失败")
		return
	}
	// 删除团队
	if _, err := teamService.Delete("id = ?", teamId); err != nil {
		common.ServerError(c, "团队删除失败")
		return
	}
	common.Success(c, "")
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

	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	var req struct {
		TeamId         string `json:"team_id" binding:"required"`
		ApplicantNotes string `json:"applicant_notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "")
		return
	}
	teamId := (req.TeamId)
	if teamId == "" {
		common.BadRequest(c, "参数错误：team_id")
		return
	}
	teamService := services.NewTeamService()
	var team models.Team
	if err := teamService.GetById(teamId, &team); err != nil {
		if errors.Is(err, services.ErrRecordNotFound) {
			failResponseData["code"] = FailCodeTeamNotExist
			common.BadRequestData(c, "团队不存在", failResponseData)
		} else {
			common.ServerError(c, "查询错误")
		}
		return
	}
	if !team.OpenInvite {
		failResponseData["code"] = FailCodeInvitedNotOpen
		common.BadRequestData(c, "团队未开启邀请", failResponseData)
		return
	}
	invitedPermType := team.InvitedPermType
	teamMemberService := teamService.TeamMemberService
	if ok, err := teamMemberService.Exist("deleted_at is null and team_id = ? and user_id = ?", teamId, userId); ok {
		failResponseData["code"] = FailCodeAlreadyJoined
		common.BadRequestData(c, "已加入团队", failResponseData)
		return
	} else if err != nil {
		common.ServerError(c, "查询错误.")
		return
	}
	teamJoinRequestService := teamService.TeamJoinRequestService
	if ok, err := teamJoinRequestService.Exist("deleted_at is null and team_id = ? and user_id = ? and status = ?", teamId, userId, models.TeamJoinRequestStatusPending); ok {
		failResponseData["code"] = FailCodeAlreadyApplied
		common.BadRequestData(c, "不能重复申请", failResponseData)
		return
	} else if err != nil {
		common.ServerError(c, "查询错误..")
		return
	}
	if err := teamJoinRequestService.Create(&models.TeamJoinRequest{
		UserId:         userId,
		TeamId:         teamId,
		PermType:       invitedPermType,
		ApplicantNotes: req.ApplicantNotes,
	}); err != nil {
		common.ServerError(c, "新建错误")
	}
	common.Success(c, "")
}

// GetTeamJoinRequestList 获取申请列表
func GetTeamJoinRequestList(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	teamId := (c.Query("team_id"))
	// if teamId <= 0 {
	// 	teamId = 0
	// }
	startTimeStr := ""
	startTimeInt := str.DefaultToInt(c.Query("start_time"), 0)
	if startTimeInt > 0 {
		startTimeStr = myTime.Time(time.UnixMilli(startTimeInt)).String()
	}
	teamService := services.NewTeamService()
	teamJoinRequestMessageShowService := teamService.TeamJoinRequestMessageShowService
	now := myTime.Time(time.Now())
	result := teamService.FindTeamJoinRequest(userId, teamId, startTimeStr)
	// 获取用户信息
	userIds := make([]string, 0)
	for _, item := range result {
		userIds = append(userIds, item.TeamJoinRequest.UserId)
	}
	userMap, err := GetUsersInfo(c, userIds)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}
	for i := range result {
		userId := result[i].TeamJoinRequest.UserId
		userInfo, exists := userMap[userId]
		if exists {
			result[i].User = &models.UserProfile{
				Id:       userInfo.UserID,
				Nickname: userInfo.Nickname,
				Avatar:   userInfo.Avatar,
			}
		}
	}
	if startTimeStr == "" {
		common.Success(c, result)
		return
	}

	var messageShowList []models.TeamJoinRequestMessageShow
	if err := teamJoinRequestMessageShowService.Find(&messageShowList, "user_id = ? and team_id = ?", userId, teamId); err != nil {
		log.Println("TeamJoinRequestMessageShow查询错误", err)
		common.ServerError(c, "查询错误")
		return
	}
	result = sliceutil.FilterT(func(item services.TeamJoinRequestQuery) bool {
		return sliceutil.Find(func(messageShowItem models.TeamJoinRequestMessageShow) bool {
			return messageShowItem.TeamJoinRequestId == item.TeamJoinRequest.Id
		}, messageShowList...) == nil
	}, result...)
	newMessageShowList := sliceutil.MapT(func(item services.TeamJoinRequestQuery) models.BaseModel {
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
			common.ServerError(c, "新建错误")
			return
		}
	}
	common.Success(c, result)
}

// GetSelfTeamJoinRequestList 获取自身的申请列表
func GetSelfTeamJoinRequestList(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	teamId := (c.Query("team_id"))
	// if teamId <= 0 {
	// 	teamId = 0
	// }
	startTimeStr := ""
	startTimeInt := str.DefaultToInt(c.Query("start_time"), 0)
	if startTimeInt > 0 {
		startTimeStr = myTime.Time(time.UnixMilli(startTimeInt)).String()
	}
	teamService := services.NewTeamService()
	result := teamService.FindSelfTeamJoinRequest(userId, teamId, startTimeStr)
	// 获取用户信息
	userIds := make([]string, 0)
	for _, item := range result {
		userIds = append(userIds, item.TeamJoinRequest.ProcessedBy)
	}

	userMap, err := GetUsersInfo(c, userIds)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}
	for i := range result {
		userId := result[i].TeamJoinRequest.ProcessedBy
		userInfo, exists := userMap[userId]
		if exists {
			result[i].User = &models.UserProfile{
				Id:       userInfo.UserID,
				Nickname: userInfo.Nickname,
				Avatar:   userInfo.Avatar,
			}
		}
	}
	common.Success(c, result)
}

// ReviewTeamJoinRequest 权限申请审核
func ReviewTeamJoinRequest(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	var req struct {
		ApplyId      string `json:"apply_id" binding:"required"`
		ApprovalCode uint8  `json:"approval_code" binding:"min=0,max=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "")
		return
	}
	teamJoinRequestsId := str.DefaultToInt(req.ApplyId, 0)
	if teamJoinRequestsId <= 0 {
		common.BadRequest(c, "参数错误：apply_id")
		return
	}
	approvalCode := req.ApprovalCode
	if approvalCode != 0 && approvalCode != 1 {
		common.BadRequest(c, "参数错误：approval_code")
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
			common.BadRequest(c, "申请已被处理或无权限")
		} else {
			common.ServerError(c, "查询错误")
		}
		return
	}
	if teamJoinRequest.PermType < models.TeamPermTypeReadOnly || teamJoinRequest.PermType > models.TeamPermTypeEditable {
		common.BadRequest(c, "参数错误：teamJoinRequest.PermType")
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
		common.ServerError(c, "更新错误")
		return
	}
	if approvalCode == 1 {
		if permType, err := teamService.GetTeamPermTypeByForUser(teamJoinRequest.TeamId, teamJoinRequest.UserId); err != nil || permType == nil {
			common.ServerError(c, "查询错误")
			return
		} else if *permType == models.TeamPermTypeNone {
			if err := teamService.TeamMemberService.Create(&models.TeamMember{
				TeamId:   teamJoinRequest.TeamId,
				UserId:   teamJoinRequest.UserId,
				PermType: teamJoinRequest.PermType,
			}); err != nil {
				common.ServerError(c, "新建错误")
				return
			}
		} else if teamJoinRequest.PermType <= *permType {
			common.Success(c, "")
			return
		} else {
			if _, err := teamService.TeamMemberService.UpdatesIgnoreZero(&models.TeamMember{
				PermType: teamJoinRequest.PermType,
			}, "team_id = ? and user_id = ?", teamJoinRequest.TeamId, teamJoinRequest.UserId); err != nil {
				common.ServerError(c, "更新错误")
				return
			}
		}
	}
	common.Success(c, "")
}

// SetTeamInfo 设置团队信息
func SetTeamInfo(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	var req struct {
		TeamId      string `json:"team_id" form:"team_id" binding:"required"`
		Name        string `json:"name" form:"name"`
		Description string `json:"description" form:"description"`
	}
	if err := c.ShouldBind(&req); err != nil {
		common.BadRequest(c, err.Error())
		return
	}
	teamId := (req.TeamId)
	if teamId == "" {
		common.BadRequest(c, "参数错误：team_id")
		return
	}
	teamService := services.NewTeamService()
	if permType, err := teamService.GetTeamPermTypeByForUser(teamId, userId); err != nil || permType == nil {
		common.ServerError(c, "查询错误")
		return
	} else if *permType < models.TeamPermTypeAdmin {
		common.Forbidden(c, "")
		return
	}
	fileHeader, err := c.FormFile("avatar")
	fileExists := err == nil
	if fileExists && fileHeader.Size > 2<<20 {
		common.BadRequest(c, "文件大小不能超过2MB")
		return
	}
	if req.Name == "" && req.Description == "" && !fileExists {
		common.BadRequest(c, "")
		return
	}
	reviewClient := services.GetSafereviewClient()
	if req.Name != "" || req.Description != "" {
		if req.Name != "" && reviewClient != nil {
			reviewResponse, err := (reviewClient).ReviewText(req.Name)
			if err != nil {
				log.Println("名称审核失败", req.Name, err)
				common.ReviewFail(c, "审核失败")
				return
			} else if reviewResponse != nil && reviewResponse.Status != safereview.ReviewTextResultPass {
				log.Println("名称审核不通过", req.Name, reviewResponse)
				common.ReviewFail(c, "审核不通过")
				return
			}
		}
		if req.Description != "" && reviewClient != nil {
			reviewResponse, err := (reviewClient).ReviewText(req.Description)
			if err != nil {
				log.Println("描述审核失败", req.Description, err)
				common.ReviewFail(c, "审核失败")
				return
			} else if reviewResponse != nil && reviewResponse.Status != safereview.ReviewTextResultPass {
				log.Println("描述审核不通过", req.Description, reviewResponse)
				common.ReviewFail(c, "审核不通过")
				return
			}
		}
		if _, err := teamService.UpdatesIgnoreZeroById(teamId, &models.Team{
			Name:        req.Name,
			Description: req.Description,
		}); err != nil {
			common.ServerError(c, "更新错误")
			return
		}
	}
	result := map[string]any{}
	if fileExists {
		file, err := fileHeader.Open()
		if err != nil {
			common.BadRequest(c, "获取文件失败")
			return
		}
		defer file.Close()
		fileBytes := make([]byte, fileHeader.Size)
		if _, err := file.Read(fileBytes); err != nil {
			common.BadRequest(c, "读取文件失败")
			return
		}
		contentType := fileHeader.Header.Get("Content-Type")
		if reviewClient != nil {
			base64Str := base64.StdEncoding.EncodeToString(fileBytes)
			reviewResponse, err := (reviewClient).ReviewPictureFromBase64(base64Str)
			if err != nil {
				log.Println("头像审核错误", err)
				common.ReviewFail(c, "头像审核错误")
				return
			} else if reviewResponse.Status != safereview.ReviewImageResultPass {
				log.Println("头像审核不通过", err, reviewResponse)
				common.ReviewFail(c, "头像审核不通过")
				return
			}
		}
		avatarPath, err := teamService.UploadTeamAvatarById(teamId, fileBytes, contentType)
		if err == nil {
			result["avatar"] = services.GetConfig().StorageUrl.Attatch + avatarPath
			// result["avatar"] = avatarPath
		} else {
			log.Println("上传头像失败", err)
		}
	}
	common.Success(c, result)
}

// SetTeamInvited 修改团队邀请设置
func SetTeamInvited(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	var req struct {
		TeamId          string               `json:"team_id" binding:"required"`
		InvitedPermType *models.TeamPermType `json:"invited_perm_type"`
		OpenInvite      *bool                `json:"open_invite"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, err.Error())
		return
	}
	teamId := (req.TeamId)
	if teamId == "" {
		common.BadRequest(c, "参数错误：team_id")
		return
	}
	if req.InvitedPermType == nil && req.OpenInvite == nil {
		common.BadRequest(c, "")
		return
	}
	if req.InvitedPermType != nil && (*req.InvitedPermType < models.TeamPermTypeReadOnly || *req.InvitedPermType > models.TeamPermTypeEditable) {
		common.BadRequest(c, "参数错误：invited_perm_type")
		return
	}
	teamService := services.NewTeamService()
	if permType, err := teamService.GetTeamPermTypeByForUser(teamId, userId); err != nil || permType == nil {
		common.ServerError(c, "查询错误")
		return
	} else if *permType < models.TeamPermTypeAdmin {
		common.Forbidden(c, "")
		return
	}
	updateColumns := map[string]any{}
	if req.InvitedPermType != nil {
		updateColumns["invited_perm_type"] = *req.InvitedPermType
	}
	if req.OpenInvite != nil {
		updateColumns["open_invite"] = *req.OpenInvite
	}
	if _, err := teamService.UpdateColumnsById(teamId, updateColumns); err != nil {
		common.ServerError(c, "更新错误")
		return
	}
	common.Success(c, "")
}

// GetTeamInvitedInfo 获取团队信息
func GetTeamInvitedInfo(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	teamId := (c.Query("team_id"))
	if teamId == "" {
		common.BadRequest(c, "参数错误：team_id")
		return
	}
	teamService := services.NewTeamService()
	var team models.Team
	if err := teamService.GetById(teamId, &team); err != nil {
		if errors.Is(err, services.ErrRecordNotFound) {
			common.BadRequest(c, "团队不存在")
		} else {
			common.ServerError(c, "查询错误")
		}
		return
	}
	selfPermType, err := teamService.GetTeamPermTypeByForUser(teamId, userId)
	if err != nil {
		common.ServerError(c, "查询错误")
		return
	}
	result := map[string]any{
		"id":                (team.Id),
		"name":              team.Name,
		"self_perm_type":    selfPermType,
		"invited_perm_type": team.InvitedPermType,
		"invited_switch":    team.OpenInvite,
	}
	common.Success(c, result)
}

// ExitTeam 退出团队
func ExitTeam(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	var req struct {
		TeamId string `json:"team_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "")
		return
	}
	teamId := (req.TeamId)
	if teamId == "" {
		common.BadRequest(c, "参数错误：team_id")
		return
	}
	teamService := services.NewTeamService()
	if permType, err := teamService.GetTeamPermTypeByForUser(teamId, userId); err != nil || permType == nil {
		common.ServerError(c, "查询错误")
		return
	} else if *permType == models.TeamPermTypeCreator {
		common.Forbidden(c, "团队创建者不能退出团队，需要先转移团队")
		return
	}
	// 退出或删除项目 todo

	// 删除团队成员
	teamMemberService := services.NewTeamMemberService()
	if _, err := teamMemberService.Delete("team_id = ? and user_id = ?", teamId, userId); err != nil {
		common.ServerError(c, "团队成员删除失败")
		return
	}
	common.Success(c, "")
}

// SetTeamMemberPermission 设置团队成员权限
func SetTeamMemberPermission(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	var req struct {
		TeamId   string               `json:"team_id" binding:"required"`
		UserId   string               `json:"user_id" binding:"required"`
		PermType *models.TeamPermType `json:"perm_type" binding:"min=0,max=2"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "")
		return
	}
	teamId := (req.TeamId)
	if teamId == "" {
		common.BadRequest(c, "参数错误：team_id")
		return
	}
	reqUserId := req.UserId
	if reqUserId == "" {
		common.BadRequest(c, "参数错误：user_id")
		return
	}
	if req.PermType == nil {
		common.BadRequest(c, "参数错误：perm_type")
		return
	}
	teamService := services.NewTeamService()
	selfPermType, err := teamService.GetTeamPermTypeByForUser(teamId, userId)
	if err != nil {
		common.ServerError(c, "查询错误")
		return
	} else if selfPermType == nil || *selfPermType < models.TeamPermTypeAdmin {
		common.Forbidden(c, "权限不足")
		return
	}
	permType, err := teamService.GetTeamPermTypeByForUser(teamId, reqUserId)
	if err != nil || permType == nil {
		common.ServerError(c, "查询错误")
		return
	}
	if *permType >= *selfPermType {
		common.Forbidden(c, "不可设置比自己更高的权限")
		return
	}
	teamMemberService := services.NewTeamMemberService()
	if _, err := teamMemberService.UpdateColumns(map[string]any{
		"perm_type": req.PermType,
	}, "team_id = ? and user_id = ?", teamId, reqUserId); err != nil {
		common.ServerError(c, "更新错误")
		return
	}
	common.Success(c, "")
}

// ChangeTeamCreator 更改团队创建者
func ChangeTeamCreator(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	var req struct {
		TeamId string `json:"team_id" binding:"required"`
		UserId string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "")
		return
	}
	teamId := (req.TeamId)
	if teamId == "" {
		common.BadRequest(c, "参数错误：team_id")
		return
	}
	reqUserId := req.UserId
	if reqUserId == "" {
		common.BadRequest(c, "参数错误：user_id")
		return
	}
	if userId == reqUserId {
		common.BadRequest(c, "不能转移给自己")
		return
	}
	teamService := services.NewTeamService()
	selfPermType, err := teamService.GetTeamPermTypeByForUser(teamId, userId)
	if err != nil {
		common.ServerError(c, "查询错误")
		return
	} else if selfPermType == nil || *selfPermType != models.TeamPermTypeCreator {
		common.Forbidden(c, "")
		return
	}
	teamMemberService := services.NewTeamMemberService()
	transactDB := teamMemberService.DBModule.DB.Begin() // 开启事务
	needRollback := false
	defer func() {
		if needRollback {
			transactDB.Rollback()
		} else {
			transactDB.Commit()
		}
	}()
	if err := transactDB.Error; err != nil {
		common.ServerError(c, "更新错误")
		needRollback = true
		return
	}
	if _, err := teamMemberService.UpdateColumns(map[string]any{
		"perm_type": models.TeamPermTypeAdmin,
	}, "team_id = ? and user_id = ?", teamId, userId); err != nil {
		common.ServerError(c, "更新错误.")
		needRollback = true
		return
	}
	if _, err := teamMemberService.UpdateColumns(map[string]any{
		"perm_type": models.TeamPermTypeCreator,
	}, "team_id = ? and user_id = ?", teamId, reqUserId); err != nil {
		common.ServerError(c, "更新错误..")
		needRollback = true
		return
	}
	common.Success(c, "")
}

// RemoveTeamMember 移除团队成员
func RemoveTeamMember(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	teamId := (c.Query("team_id"))
	if teamId == "" {
		common.BadRequest(c, "参数错误：team_id")
		return
	}
	reqUserId := c.Query("user_id")
	if reqUserId == "" {
		common.BadRequest(c, "参数错误：user_id")
		return
	}
	if userId == reqUserId {
		common.BadRequest(c, "不能移除自己")
		return
	}
	teamService := services.NewTeamService()
	selfPermType, err := teamService.GetTeamPermTypeByForUser(teamId, userId)
	if err != nil {
		common.ServerError(c, "查询错误")
		return
	} else if selfPermType == nil || *selfPermType < models.TeamPermTypeAdmin {
		common.Forbidden(c, "")
		return
	}
	permType, err := teamService.GetTeamPermTypeByForUser(teamId, reqUserId)
	if err != nil || permType == nil {
		common.ServerError(c, "查询错误")
		return
	} else if *permType >= *selfPermType {
		common.Forbidden(c, "")
		return
	}
	// 退出或删除项目 todo

	// 删除团队成员
	teamMemberService := services.NewTeamMemberService()
	if _, err := teamMemberService.Delete("team_id = ? and user_id = ?", teamId, reqUserId); err != nil {
		common.ServerError(c, "团队成员删除失败")
		return
	}
	common.Success(c, "")
}

// SetTeamMemberNickname 设置团队成员昵称
func SetTeamMemberNickname(c *gin.Context) {
	userId, err := utils.GetUserId(c)
	if err != nil {
		common.Unauthorized(c)
		return
	}
	var req struct {
		TeamId   string `json:"team_id" binding:"required"`
		UserId   string `json:"user_id" binding:"required"`
		Nickname string `json:"nickname" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "")
		return
	}
	teamId := (req.TeamId)
	if teamId == "" {
		common.BadRequest(c, "参数错误：team_id")
		return
	}
	reqUserId := req.UserId
	if reqUserId == "" {
		common.BadRequest(c, "参数错误：user_id")
		return
	}
	teamService := services.NewTeamService()
	teamMemberService := teamService.TeamMemberService
	if ok, err := teamMemberService.Exist("team_id = ? and user_id = ?", teamId, reqUserId); err != nil {
		common.ServerError(c, "查询错误")
		return
	} else if !ok {
		common.BadRequest(c, "团队成员不存在")
		return
	}
	if userId != reqUserId {
		selfPermType, err := teamService.GetTeamPermTypeByForUser(teamId, userId)
		if err != nil {
			common.ServerError(c, "查询错误")
			return
		} else if selfPermType == nil || *selfPermType < models.TeamPermTypeAdmin {
			common.Forbidden(c, "")
			return
		}
	}
	reviewClient := services.GetSafereviewClient()
	if reviewClient != nil {
		reviewResponse, err := (reviewClient).ReviewText(req.Nickname)
		if err != nil {
			log.Println("昵称审核失败", req.Nickname, err)
			common.ReviewFail(c, "审核失败")
			return
		} else if reviewResponse != nil && reviewResponse.Status != safereview.ReviewTextResultPass {
			log.Println("昵称审核不通过", req.Nickname, reviewResponse)
			common.ReviewFail(c, "审核不通过")
			return
		}
	}

	if _, err := teamMemberService.UpdateColumns(map[string]any{
		"nickname": req.Nickname,
	}, "team_id = ? and user_id = ?", teamId, reqUserId); err != nil {
		common.ServerError(c, "更新错误")
		return
	}
	common.Success(c, "")
}
