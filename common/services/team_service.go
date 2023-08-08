package services

import (
	"errors"
	"fmt"
	"io"
	"protodesign.cn/kcserver/common"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/storage"
	storageBase "protodesign.cn/kcserver/utils/storage/base"
	"protodesign.cn/kcserver/utils/str"
	"protodesign.cn/kcserver/utils/time"
	"strings"
)

type TeamService struct {
	*DefaultService
	TeamMemberService      *TeamMemberService
	TeamJoinRequestService *TeamJoinRequestService
}

func NewTeamService() *TeamService {
	that := &TeamService{
		DefaultService:         NewDefaultService(&models.Team{}),
		TeamMemberService:      NewTeamMemberService(),
		TeamJoinRequestService: NewTeamJoinRequestService(),
	}
	that.That = that
	return that
}

type TeamMemberService struct {
	*DefaultService
}

func NewTeamMemberService() *TeamMemberService {
	that := &TeamMemberService{
		DefaultService: NewDefaultService(&models.TeamMember{}),
	}
	that.That = that
	return that
}

type TeamJoinRequestService struct {
	*DefaultService
}

func NewTeamJoinRequestService() *TeamJoinRequestService {
	that := &TeamJoinRequestService{
		DefaultService: NewDefaultService(&models.TeamJoinRequest{}),
	}
	that.That = that
	return that
}

func (s *TeamService) UploadTeamAvatar(team *models.Team, file io.Reader, fileSize int64, contentType string) (string, error) {
	if fileSize > 1024*1024*5 {
		return "", errors.New("文件大小不能超过5MB")
	}
	var suffix string
	switch contentType {
	case "image/jpeg":
		suffix = "jpg"
	case "image/png":
		suffix = "png"
	case "image/gif":
		suffix = "gif"
	case "image/bmp":
		suffix = "bmp"
	case "image/tiff":
		suffix = "tif"
	case "image/webp":
		suffix = "webp"
	default:
		return "", errors.New(fmt.Sprintf("不支持的文件类型：%s", contentType))
	}
	if team.Uid == "" {
		team.Uid = str.GetUid()
	}
	fileName := fmt.Sprintf("%s.%s", str.GetUid(), suffix)
	avatarPath := fmt.Sprintf("/teams/%s/avatar/%s", team.Uid, fileName)
	if _, err := storage.FilesBucket.PutObject(&storageBase.PutObjectInput{
		ObjectName:  avatarPath,
		Reader:      file,
		ObjectSize:  fileSize,
		ContentType: contentType,
	}); err != nil {
		return "", errors.New("上传文件失败")
	}
	team.Avatar = avatarPath
	if s.UpdateColumnsById(team.Id, map[string]any{
		"avatar": avatarPath,
	}) != nil {
		return "", errors.New("更新错误")
	}
	return avatarPath, nil
}

type Team struct {
	Id              int64               `json:"id"`
	CreatedAt       time.Time           `json:"created_at"`
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	Avatar          string              `json:"avatar"`
	InvitedPermType models.TeamPermType `json:"invited_perm_type"`
}

func (team *Team) MarshalJSON() ([]byte, error) {
	if strings.HasPrefix(team.Avatar, "/") {
		team.Avatar = common.StorageHost + team.Avatar
	}
	return models.MarshalJSON(team)
}

type TeamJoinRequest models.TeamJoinRequest

func (model *TeamJoinRequest) MarshalJSON() ([]byte, error) {
	return models.MarshalJSON(model)
}

type TeamMember models.TeamMember

func (model *TeamMember) MarshalJSON() ([]byte, error) {
	return models.MarshalJSON(model)
}

type TeamJoinRequestsQueryResItem struct {
	Team            Team            `gorm:"embedded;embeddedPrefix:team__" json:"team" table:"team" join:"team;inner;id,team_id"`
	User            User            `gorm:"embedded;embeddedPrefix:user__" json:"user" table:"user" join:"user;inner;id,user_id"`
	TeamMember      TeamMember      `gorm:"-" json:"-" table:"team_member" join:"team_member;inner;team_id,team_id;user_id,user_id"`
	TeamJoinRequest TeamJoinRequest `gorm:"embedded;embeddedPrefix:team_join_request__" json:"request" table:"team_join_request"`
}

// FindTeamJoinRequests 获取用户所创建或担任管理员的团队的加入申请列表
func (s *TeamService) FindTeamJoinRequests(userId int64, teamId int64, startTime string) []TeamJoinRequestsQueryResItem {
	var result []TeamJoinRequestsQueryResItem
	whereArgsList := []WhereArgs{{Query: "team_member.user_id = ? and (team_member.perm_type = ? or team_member.perm_type = ?)", Args: []any{userId, models.TeamPermTypeAdmin, models.TeamPermTypeCreator}}}
	if teamId != 0 {
		whereArgsList = append(whereArgsList, WhereArgs{Query: "team.id = ?", Args: []any{teamId}})
	}
	if startTime != "" {
		whereArgsList = append(whereArgsList, WhereArgs{
			Query: "team_join_request.status = ? and team_join_request.created_at >= ? and team_join_request.first_displayed_at is null",
			Args:  []any{models.TeamJoinRequestStatusPending, startTime}},
		)
	}
	_ = s.TeamJoinRequestService.Find(
		&result,
		whereArgsList,
		&OrderLimitArgs{"team_join_request.id desc", 0},
	)
	return result
}
