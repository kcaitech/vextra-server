package services

import (
	"errors"
	"sort"

	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/utils/sliceutil"
)

type ProjectService struct {
	*DefaultService
	ProjectMemberService                 *ProjectMemberService
	ProjectJoinRequestService            *ProjectJoinRequestService
	ProjectJoinRequestMessageShowService *ProjectJoinRequestMessageShowService
	ProjectFavoriteService               *ProjectFavoriteService
}

func NewProjectService() *ProjectService {
	that := &ProjectService{
		DefaultService:                       NewDefaultService(&models.Project{}),
		ProjectMemberService:                 NewProjectMemberService(),
		ProjectJoinRequestService:            NewProjectJoinRequestService(),
		ProjectJoinRequestMessageShowService: NewProjectJoinRequestMessageShowService(),
		ProjectFavoriteService:               NewProjectFavoriteService(),
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

func NewProjectJoinRequestService() *ProjectJoinRequestService {
	that := &ProjectJoinRequestService{
		DefaultService: NewDefaultService(&models.ProjectJoinRequest{}),
	}
	that.That = that
	return that
}

type ProjectJoinRequestMessageShowService struct {
	*DefaultService
}

func NewProjectJoinRequestMessageShowService() *ProjectJoinRequestMessageShowService {
	that := &ProjectJoinRequestMessageShowService{
		DefaultService: NewDefaultService(&models.ProjectJoinRequestMessageShow{}),
	}
	that.That = that
	return that
}

type ProjectFavoriteService struct {
	*DefaultService
}

func NewProjectFavoriteService() *ProjectFavoriteService {
	that := &ProjectFavoriteService{
		DefaultService: NewDefaultService(&models.ProjectFavorite{}),
	}
	that.That = that
	return that
}

// type Project struct {
// 	Id            int64                  `json:"id"`
// 	CreatedAt     time.Time              `json:"created_at"`
// 	TeamId        int64                  `json:"team_id"`
// 	Name          string                 `json:"name"`
// 	Description   string                 `json:"description"`
// 	PublicSwitch  bool                   `json:"public_switch"`
// 	PermType      models.ProjectPermType `json:"perm_type"`
// 	InvitedSwitch bool                   `json:"invited_switch"`
// 	NeedApproval  bool                   `json:"need_approval"`
// }

// func (project Project) MarshalJSON() ([]byte, error) {
// 	return models.MarshalJSON(project)
// }

// type ProjectMember models.ProjectMember

// func (model ProjectMember) MarshalJSON() ([]byte, error) {
// 	return models.MarshalJSON(model)
// }

// type ProjectJoinRequest models.ProjectJoinRequest

// func (model ProjectJoinRequest) MarshalJSON() ([]byte, error) {
// 	return models.MarshalJSON(model)
// }

// GetProjectPermTypeByForUser 获取用户在项目中的权限
func (s *ProjectService) GetProjectPermTypeByForUser(projectId string, userId string) (*models.ProjectPermType, error) {
	var projectMember models.ProjectMember
	err := s.ProjectMemberService.Get(&projectMember, WhereArgs{Query: "project_id = ? and user_id = ?", Args: []any{projectId, userId}})
	if err == nil {
		return &projectMember.PermType, nil
	}
	if !errors.Is(err, ErrRecordNotFound) {
		return nil, err
	}
	type ProjectQuery struct {
		models.BaseModelStruct
		Project    models.Project    `gorm:"embedded;embeddedPrefix:p__" table:"p"`
		TeamMember models.TeamMember `gorm:"embedded;embeddedPrefix:tm__" join:",tm;inner;team_id,team_id;user_id,?user_id"`
	}
	var projectQueryResult ProjectQuery
	whereArgsList := []WhereArgs{
		{"tm.deleted_at is null", nil},
		{"p.id = ? and p.open_invite = ?", []any{projectId, true}},
	}
	err = s.Get(
		&projectQueryResult,
		&As{BaseService: s, Alias: "p"},
		&ParamArgs{"?user_id": userId},
		&whereArgsList,
	)
	if err == nil {
		return &projectQueryResult.Project.PermType, nil
	}
	if !errors.Is(err, ErrRecordNotFound) {
		return nil, err
	}
	permNone := models.ProjectPermTypeNone
	return &permNone, nil
}

type ProjectQuery struct {
	Project              models.Project       `gorm:"embedded;embeddedPrefix:p__" json:"project" table:"p"`
	CreatorProjectMember models.ProjectMember `gorm:"embedded;embeddedPrefix:pm__" json:"-" join:"project_member,pm;inner;project_id,id;perm_type,?creator_perm_type"`
	// CreatorUser          string                 `gorm:"embedded;embeddedPrefix:u__" json:"creator"`
	CreatorTeamMember   models.TeamMember      `gorm:"embedded;embeddedPrefix:ctm__" json:"-" join:"team_member,ctm;left;team_id;user_id,pm.user_id"`
	CreatorTeamNickname string                 `gorm:"-" json:"creator_team_nickname"`
	SelfPermType        models.ProjectPermType `gorm:"-" json:"self_perm_type"`
	IsInTeam            bool                   `gorm:"-" json:"is_in_team"`
	IsInvited           bool                   `gorm:"-" json:"is_invited"`
}

type SelfProjectQuery struct { // 通过邀请进入的项目
	ProjectQuery
	SelfProjectMember models.ProjectMember `gorm:"embedded;embeddedPrefix:pm1__" json:"-" join:"project_member,pm1;inner;project_id,id;user_id,?user_id"`
	// SelfUser          string               `gorm:"embedded;embeddedPrefix:u1__" json:"-" join:"user,u1;inner;id,pm1.user_id"`
	SelfTeamMember models.TeamMember `gorm:"embedded;embeddedPrefix:tm__" join:"team_member,tm;left;deleted_at,##is null;team_id,team_id;user_id,?user_id"`
}

type PublishProjectQuery struct { // 通过项目的团队公开权限进入的项目
	ProjectQuery
	TeamMember models.TeamMember `gorm:"embedded;embeddedPrefix:tm__" join:",tm;inner;team_id,team_id;user_id,?user_id"`
}

// 查询用户的项目列表
func (s *ProjectService) findProject(teamId string, userId string, projectIdList *[]string) []*ProjectQuery {
	var whereArgsList []WhereArgs
	if teamId != "" {
		whereArgsList = append(whereArgsList, WhereArgs{"p.team_id = ?", []any{teamId}})
	}
	if projectIdList != nil {
		whereArgsList = append(whereArgsList, WhereArgs{"p.id in ?", []any{*projectIdList}})
	}

	var selfProjectQueryResult []SelfProjectQuery
	whereArgsList1 := append(
		whereArgsList,
		WhereArgs{"pm.deleted_at is null and pm1.deleted_at is null", nil},
	)
	_ = s.Find(
		&selfProjectQueryResult,
		&As{BaseService: s, Alias: "p"},
		&ParamArgs{"?creator_perm_type": models.ProjectPermTypeCreator, "?user_id": userId},
		&whereArgsList1,
		&OrderLimitArgs{"pm1.id desc", 0},
	)

	var publishProjectQueryResult []PublishProjectQuery
	selfProjectIdList := sliceutil.MapT(func(item SelfProjectQuery) string {
		return item.Project.Id
	}, selfProjectQueryResult...)
	whereArgsList2 := append(
		whereArgsList,
		WhereArgs{"pm.deleted_at is null and tm.deleted_at is null", nil},
		WhereArgs{"p.open_invite = ?", []any{true}},
	)
	if len(selfProjectIdList) > 0 {
		whereArgsList2 = append(whereArgsList2, WhereArgs{"p.id not in ?", []any{selfProjectIdList}})
	}
	_ = s.Find(
		&publishProjectQueryResult,
		&As{BaseService: s, Alias: "p"},
		&ParamArgs{"?creator_perm_type": models.ProjectPermTypeCreator, "?user_id": userId},
		&whereArgsList2,
		&OrderLimitArgs{"p.id desc", 0},
	)

	selfProjectQueryResultLen := len(selfProjectQueryResult)
	publishProjectQueryResultLen := len(publishProjectQueryResult)
	result := make([]*ProjectQuery, selfProjectQueryResultLen+publishProjectQueryResultLen)
	for i := range selfProjectQueryResult {
		selfProjectQueryResult[i].SelfPermType = selfProjectQueryResult[i].SelfProjectMember.PermType
		selfProjectQueryResult[i].IsInTeam = selfProjectQueryResult[i].SelfTeamMember.Id > 0
		selfProjectQueryResult[i].IsInvited = true
		result[i] = &selfProjectQueryResult[i].ProjectQuery
		result[i].CreatorTeamNickname = result[i].CreatorTeamMember.Nickname
	}
	for i := range publishProjectQueryResult {
		publishProjectQueryResult[i].SelfPermType = publishProjectQueryResult[i].Project.PermType
		publishProjectQueryResult[i].IsInTeam = true
		publishProjectQueryResult[i].IsInvited = false
		result[i+selfProjectQueryResultLen] = &publishProjectQueryResult[i].ProjectQuery
		result[i+selfProjectQueryResultLen].CreatorTeamNickname = result[i+selfProjectQueryResultLen].CreatorTeamMember.Nickname
	}

	return result
}

// FindProject 查询用户的项目列表
func (s *ProjectService) FindProject(teamId string, userId string) []*ProjectQuery {
	return s.findProject(teamId, userId, nil)
}

// FindFavorProject 查询用户收藏的项目列表
func (s *ProjectService) FindFavorProject(teamId string, userId string) []*ProjectQuery {
	var projectFavoriteList []models.ProjectFavorite
	_ = s.ProjectFavoriteService.Find(
		&projectFavoriteList,
		&WhereArgs{"user_id = ? and is_favor = true", []any{userId}},
		&OrderLimitArgs{"id desc", 0},
	)
	projectIdList := sliceutil.MapT(func(item models.ProjectFavorite) string {
		return item.ProjectId
	}, projectFavoriteList...)
	projectIdToIndex := make(map[string]int, len(projectIdList))
	for i, projectId := range projectIdList {
		projectIdToIndex[projectId] = i
	}
	favorProjectQueryResult := s.findProject(teamId, userId, &projectIdList)
	sort.Slice(favorProjectQueryResult, func(i, j int) bool {
		return projectIdToIndex[favorProjectQueryResult[i].Project.Id] < projectIdToIndex[favorProjectQueryResult[j].Project.Id]
	})
	return favorProjectQueryResult
}

type ProjectMemberQuery struct {
	ProjectMember models.ProjectMember `gorm:"embedded;embeddedPrefix:project_member__" json:"-" table:""`
	Project       models.Project       `gorm:"embedded;embeddedPrefix:project__" json:"-" join:";inner;id,project_id"`
	// User          string                 `gorm:"embedded;embeddedPrefix:user__" json:"user"`
	// PermType      models.ProjectPermType `gorm:"-" json:"perm_type"`
}

// FindProjectMember 查询某个项目中的成员列表
func (s *ProjectService) FindProjectMember(projectId string) []ProjectMemberQuery {
	var result []ProjectMemberQuery
	whereArgsList := []WhereArgs{
		{"project.deleted_at is null", nil},
		{"project_member.project_id = ?", []any{projectId}},
	}
	_ = s.ProjectMemberService.Find(
		&result,
		&whereArgsList,
		&OrderLimitArgs{"project_member.perm_type desc, project_member.id asc", 0},
	)
	// for i := range result {
	// 	result[i].PermType = result[i].ProjectMember.PermType
	// }
	return result
}

type BaseProjectJoinRequestQuery struct {
	Project            models.Project            `gorm:"embedded;embeddedPrefix:project__" json:"project" join:";inner;id,project_id"`
	ProjectJoinRequest models.ProjectJoinRequest `gorm:"embedded;embeddedPrefix:project_join_request__" json:"request" table:""`
}

type SelfProjectJoinRequestQuery struct {
	BaseProjectJoinRequestQuery
	// User string `gorm:"embedded;embeddedPrefix:user__" json:"approver" join:";inner;id,processed_by"`
}

type ProjectJoinRequestQuery struct {
	BaseProjectJoinRequestQuery
	ProjectMember models.ProjectMember `gorm:"-" json:"-" join:";inner;project_id,project_id;user_id,?user_id"` // 自己的（非申请人的）权限
	// User          string               `gorm:"embedded;embeddedPrefix:user__" json:"user"`
}

// FindProjectJoinRequest 获取用户所创建或担任管理员的项目的加入申请列表
func (s *ProjectService) FindProjectJoinRequest(userId string, projectId string, startTime string) []ProjectJoinRequestQuery {
	var result = make([]ProjectJoinRequestQuery, 0)
	whereArgsList := []WhereArgs{
		{
			Query: "project_member.deleted_at is null and project.deleted_at is null and user.deleted_at is null",
		},
		{
			Query: "project_member.perm_type >= ? and project_member.perm_type <= ?",
			Args:  []any{models.ProjectPermTypeAdmin, models.ProjectPermTypeCreator},
		},
	}
	if projectId != "" {
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

// FindSelfProjectJoinRequest 获取用户自身的项目加入申请列表
func (s *ProjectService) FindSelfProjectJoinRequest(userId string, projectId string, startTime string) []SelfProjectJoinRequestQuery {
	var result []SelfProjectJoinRequestQuery
	whereArgsList := []WhereArgs{
		{
			Query: "project.deleted_at is null",
		},
		{
			Query: "project_join_request.user_id = ? and project_join_request.status != ?",
			Args:  []any{userId, models.ProjectJoinRequestStatusPending},
		},
	}
	if projectId != "" {
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
		whereArgsList,
		&OrderLimitArgs{"project_join_request.id desc", 0},
	)
	return result
}

// ToggleProjectFavorite 收藏/取消收藏项目
func (s *ProjectService) ToggleProjectFavorite(userId string, projectId string, isFavor bool) error {
	if _, err := s.ProjectFavoriteService.HardDelete("user_id = ? and project_id = ?", userId, projectId); err != nil && !errors.Is(err, ErrRecordNotFound) {
		return err
	}
	if isFavor {
		if err := s.ProjectFavoriteService.Create(&models.ProjectFavorite{UserId: userId, ProjectId: projectId}); err != nil {
			return err
		}
	}
	return nil
}
