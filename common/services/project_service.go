package services

import (
	"errors"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/utils/time"
)

type ProjectService struct {
	*DefaultService
	ProjectMemberService      *ProjectMemberService
	ProjectJoinRequestService *ProjectJoinRequestService
}

func NewProjectService() *ProjectService {
	that := &ProjectService{
		DefaultService:            NewDefaultService(&models.Project{}),
		ProjectMemberService:      NewProjectMemberService(),
		ProjectJoinRequestService: NewprojectJoinRequestService(),
	}
	that.That = that
	return that
}

type ProjectMemberService struct {
	*DefaultService
}

func NewProjectMemberService() *ProjectMemberService {
	that := &ProjectMemberService{
		DefaultService: NewDefaultService(&models.ProjectMember{}),
	}
	that.That = that
	return that
}

type ProjectJoinRequestService struct {
	*DefaultService
}

func NewprojectJoinRequestService() *ProjectJoinRequestService {
	that := &ProjectJoinRequestService{
		DefaultService: NewDefaultService(&models.ProjectJoinRequest{}),
	}
	that.That = that
	return that
}

// GetProjectPermTypeByForUser 获取用户在项目中的权限
func (s *ProjectService) GetProjectPermTypeByForUser(projectId int64, userId int64) (*models.ProjectPermType, error) {
	var projectMember models.ProjectMember
	if err := s.ProjectMemberService.Get(&projectMember, WhereArgs{Query: "project_id = ? and user_id = ?", Args: []any{projectId, userId}}); err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &projectMember.PermType, nil
}

type Project struct {
	Id            int64                  `json:"id"`
	CreatedAt     time.Time              `json:"created_at"`
	TeamId        int64                  `json:"team_id"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	PublicSwitch  bool                   `json:"public_switch"`
	PermType      models.ProjectPermType `json:"perm_type"`
	InvitedSwitch bool                   `json:"invited_switch"`
	NeedApproval  bool                   `json:"need_approval"`
}

func (project Project) MarshalJSON() ([]byte, error) {
	return models.MarshalJSON(project)
}

type ProjectMember models.ProjectMember

func (model ProjectMember) MarshalJSON() ([]byte, error) {
	return models.MarshalJSON(model)
}

type ProjectQueryResItem struct {
	Project       Project                `gorm:"embedded;embeddedPrefix:project__" json:"project" table:"project"`
	ProjectMember ProjectMember          `gorm:"embedded;embeddedPrefix:project_member__" json:"-" table:"project_member" join:"project_member;inner;project_id,id"`
	User          User                   `gorm:"embedded;embeddedPrefix:user__" json:"-" table:"user" join:"user;inner;id,project_member.user_id"`
	PermType      models.ProjectPermType `gorm:"-" json:"perm_type"`
}

// FindProjectByTeamIdAndUserId 查询某个用户（所在的某个团队中）的所有项目列表
func (s *ProjectService) FindProjectByTeamIdAndUserId(teamId int64, userId int64) []ProjectQueryResItem {
	var result []ProjectQueryResItem
	whereArgsList := []WhereArgs{
		{"project_member.deleted_at is null and user.deleted_at is null", nil},
		{"project_member.user_id = ?", []any{userId}},
	}
	if teamId > 0 {
		whereArgsList = append(whereArgsList, WhereArgs{"project.team_id = ?", []any{teamId}})
	}
	_ = s.Find(
		&result,
		&whereArgsList,
		&OrderLimitArgs{"project_member.id desc", 0},
	)
	for i := range result {
		result[i].PermType = result[i].ProjectMember.PermType
	}
	return result
}

type ProjectMemberQueryResItem struct {
	ProjectMember ProjectMember          `gorm:"embedded;embeddedPrefix:project_member__" json:"-" table:"project_member"`
	Project       Project                `gorm:"embedded;embeddedPrefix:project__" json:"-" table:"project" join:"project;inner;id,project_id"`
	User          User                   `gorm:"embedded;embeddedPrefix:user__" json:"user" table:"user" join:"user;inner;id,user_id"`
	PermType      models.ProjectPermType `gorm:"-" json:"perm_type"`
}

// FindProjectMember 查询某个项目中的成员列表
func (s *ProjectService) FindProjectMember(projectId int64) []ProjectMemberQueryResItem {
	var result []ProjectMemberQueryResItem
	whereArgsList := []WhereArgs{
		{"project.deleted_at is null and user.deleted_at is null", nil},
		{"project_member.project_id = ?", []any{projectId}},
	}
	_ = s.ProjectMemberService.Find(
		&result,
		&whereArgsList,
		&OrderLimitArgs{"project_member.perm_type desc, project_member.id asc", 0},
	)
	for i := range result {
		result[i].PermType = result[i].ProjectMember.PermType
	}
	return result
}

type ProjectJoinRequest models.ProjectJoinRequest

func (model ProjectJoinRequest) MarshalJSON() ([]byte, error) {
	return models.MarshalJSON(model)
}

type projectJoinRequestQueryResItem struct {
	ProjectMember      ProjectMember      `gorm:"-" json:"-" table:"project_member" join:"project_member;inner;project_id,project_id;user_id,?user_id"` // 自己的（非申请人的）权限
	Project            Project            `gorm:"embedded;embeddedPrefix:project__" json:"project" table:"project" join:"project;inner;id,project_id"`
	User               User               `gorm:"embedded;embeddedPrefix:user__" json:"user" table:"user" join:"user;inner;id,user_id"`
	ProjectJoinRequest ProjectJoinRequest `gorm:"embedded;embeddedPrefix:project_join_request__" json:"request" table:"project_join_request"`
}

// FindProjectJoinRequest 获取用户所创建或担任管理员的项目的加入申请列表
func (s *ProjectService) FindProjectJoinRequest(userId int64, projectId int64, startTime string) []projectJoinRequestQueryResItem {
	var result []projectJoinRequestQueryResItem
	whereArgsList := []WhereArgs{
		{
			Query: "project_member.deleted_at is null and project.deleted_at is null and user.deleted_at is null",
		},
		{
			Query: "project_member.perm_type >= ? and project_member.perm_type <= ? and project_join_request.status = ?",
			Args:  []any{models.ProjectPermTypeAdmin, models.ProjectPermTypeCreator, models.ProjectJoinRequestStatusPending},
		},
	}
	if projectId != 0 {
		whereArgsList = append(whereArgsList, WhereArgs{Query: "project_join_request.project_id = ?", Args: []any{projectId}})
	}
	if startTime != "" {
		whereArgsList = append(whereArgsList, WhereArgs{
			Query: "project_join_request.created_at >= ? and project_join_request.first_displayed_at is null",
			Args:  []any{startTime},
		})
	}
	_ = s.ProjectJoinRequestService.Find(
		&result,
		&ParamArgs{"?user_id": userId},
		whereArgsList,
		&OrderLimitArgs{"project_join_request.id desc", 0},
	)
	return result
}
