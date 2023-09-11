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
	if _, err := s.UpdateColumnsById(team.Id, map[string]any{
		"avatar": avatarPath,
	}); err != nil {
		return "", errors.New("更新错误")
	}
	return avatarPath, nil
}

func (s *TeamService) UploadTeamAvatarById(teamId int64, file io.Reader, fileSize int64, contentType string) (string, error) {
	team := models.Team{}
	if err := s.GetById(teamId, &team); err != nil {
		return "", err
	}
	return s.UploadTeamAvatar(&team, file, fileSize, contentType)
}

// GetTeamPermTypeByForUser 获取用户在团队中的权限
func (s *TeamService) GetTeamPermTypeByForUser(teamId int64, userId int64) (*models.TeamPermType, error) {
	var teamMember TeamMember
	if err := s.TeamMemberService.Get(&teamMember, WhereArgs{Query: "team_id = ? and user_id = ?", Args: []any{teamId, userId}}); err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &teamMember.PermType, nil
}

type Team struct {
	Id              int64               `json:"id"`
	CreatedAt       time.Time           `json:"created_at"`
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	Avatar          string              `json:"avatar"`
	InvitedPermType models.TeamPermType `json:"invited_perm_type"`
	InvitedSwitch   bool                `json:"invited_switch"`
}

func (team Team) MarshalJSON() ([]byte, error) {
	if strings.HasPrefix(team.Avatar, "/") {
		team.Avatar = common.StorageHost + team.Avatar
	}
	return models.MarshalJSON(team)
}

type TeamMember models.TeamMember

func (model TeamMember) MarshalJSON() ([]byte, error) {
	return models.MarshalJSON(model)
}

type TeamQueryResItem struct {
	Team              Team                `gorm:"embedded;embeddedPrefix:t__" json:"team" table:"t"`
	SelfTeamMember    TeamMember          `gorm:"embedded;embeddedPrefix:tm__" json:"-" join:"team_member,tm;inner;team_id,id"`
	SelfUser          User                `gorm:"embedded;embeddedPrefix:u__" json:"-" join:"user,u;inner;id,tm.user_id"`
	CreatorTeamMember TeamMember          `gorm:"embedded;embeddedPrefix:tm1__" json:"-" join:"team_member,tm1;inner;team_id,id;perm_type,?creator_perm_type"`
	CreatorUser       User                `gorm:"embedded;embeddedPrefix:u1__" json:"creator" join:"user,u1;inner;id,tm1.user_id"`
	SelfPermType      models.TeamPermType `gorm:"-" json:"self_perm_type"`
}

// FindTeamByUserId 查询某个用户所在的所有团队列表
func (s *TeamService) FindTeamByUserId(userId int64) []TeamQueryResItem {
	var result []TeamQueryResItem
	whereArgsList := []WhereArgs{
		{"tm.deleted_at is null and u.deleted_at is null", nil},
		{"tm.user_id = ?", []any{userId}},
	}
	_ = s.Find(
		&result,
		&As{BaseService: s, Alias: "t"},
		&ParamArgs{"?creator_perm_type": models.TeamPermTypeCreator},
		&whereArgsList,
		&OrderLimitArgs{"tm.id desc", 0},
	)
	for i := range result {
		result[i].SelfPermType = result[i].SelfTeamMember.PermType
	}
	return result
}

type TeamMemberQueryResItem struct {
	TeamMember TeamMember          `gorm:"embedded;embeddedPrefix:team_member__" json:"-" table:""`
	Team       Team                `gorm:"embedded;embeddedPrefix:team__" json:"-" join:";inner;id,team_id"`
	User       User                `gorm:"embedded;embeddedPrefix:user__" json:"user" join:";inner;id,user_id"`
	PermType   models.TeamPermType `gorm:"-" json:"perm_type"`
}

// FindTeamMember 查询某个团队的成员列表
func (s *TeamService) FindTeamMember(teamId int64) []TeamMemberQueryResItem {
	var result []TeamMemberQueryResItem
	whereArgsList := []WhereArgs{
		{"team.deleted_at is null and user.deleted_at is null", nil},
		{"team_member.team_id = ?", []any{teamId}},
	}
	_ = s.TeamMemberService.Find(
		&result,
		&whereArgsList,
		&OrderLimitArgs{"team_member.perm_type desc, team_member.id asc", 0},
	)
	for i := range result {
		result[i].PermType = result[i].TeamMember.PermType
	}
	return result
}

type TeamJoinRequest models.TeamJoinRequest

func (model TeamJoinRequest) MarshalJSON() ([]byte, error) {
	return models.MarshalJSON(model)
}

type SelfTeamJoinRequestsQueryResItem struct {
	Team            Team            `gorm:"embedded;embeddedPrefix:team__" json:"team" join:";inner;id,team_id"`
	TeamJoinRequest TeamJoinRequest `gorm:"embedded;embeddedPrefix:team_join_request__" json:"request" table:""`
}

type TeamJoinRequestsQueryResItem struct {
	SelfTeamJoinRequestsQueryResItem
	TeamMember TeamMember `gorm:"-" json:"-" join:";inner;team_id,team_id;user_id,?user_id"` // 自己的（非申请人的）权限
	User       User       `gorm:"embedded;embeddedPrefix:user__" json:"user" join:";inner;id,user_id"`
}

// FindTeamJoinRequest 获取用户所创建或担任管理员的团队的加入申请列表
func (s *TeamService) FindTeamJoinRequest(userId int64, teamId int64, startTime string) []TeamJoinRequestsQueryResItem {
	var result []TeamJoinRequestsQueryResItem
	whereArgsList := []WhereArgs{
		{
			Query: "team_member.deleted_at is null and team.deleted_at is null and user.deleted_at is null",
		},
		{
			Query: "team_member.perm_type >= ? and team_member.perm_type <= ?",
			Args:  []any{models.TeamPermTypeAdmin, models.TeamPermTypeCreator},
		},
	}
	if teamId != 0 {
		whereArgsList = append(whereArgsList, WhereArgs{Query: "team_join_request.team_id = ?", Args: []any{teamId}})
	}
	if startTime != "" {
		whereArgsList = append(whereArgsList, WhereArgs{
			Query: "team_join_request.created_at >= ? and team_join_request.first_displayed_at is null",
			Args:  []any{startTime}},
		)
	}
	_ = s.TeamJoinRequestService.Find(
		&result,
		&ParamArgs{"?user_id": userId},
		whereArgsList,
		&OrderLimitArgs{"team_join_request.id desc", 0},
	)
	return result
}

// FindSelfTeamJoinRequest 获取用户自身的团队加入申请列表
func (s *TeamService) FindSelfTeamJoinRequest(userId int64, teamId int64, startTime string) []SelfTeamJoinRequestsQueryResItem {
	var result []SelfTeamJoinRequestsQueryResItem
	whereArgsList := []WhereArgs{
		{
			Query: "team.deleted_at is null",
		},
		{
			Query: "team_join_request.user_id = ? and team_join_request.status != ?",
			Args:  []any{userId, models.TeamJoinRequestStatusPending},
		},
	}
	if teamId != 0 {
		whereArgsList = append(whereArgsList, WhereArgs{Query: "team_join_request.team_id = ?", Args: []any{teamId}})
	}
	if startTime != "" {
		whereArgsList = append(whereArgsList, WhereArgs{
			Query: "team_join_request.created_at >= ? and team_join_request.first_displayed_at is null",
			Args:  []any{startTime}},
		)
	}
	_ = s.TeamJoinRequestService.Find(
		&result,
		whereArgsList,
		&OrderLimitArgs{"team_join_request.id desc", 0},
	)
	return result
}
