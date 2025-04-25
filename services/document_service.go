package services

import (
	"errors"
	"log"
	"time"

	"kcaitech.com/kcserver/models"
	"kcaitech.com/kcserver/utils/math"
	"kcaitech.com/kcserver/utils/sliceutil"
)

type DocumentService struct {
	*DefaultService
	DocumentPermissionService         *DocumentPermissionService
	DocumentAccessRecordService       *DocumentAccessRecordService
	DocumentFavoritesService          *DocumentFavoritesService
	DocumentPermissionRequestsService *DocumentPermissionRequestsService
	DocumentVersionService            *DocumentVersionService
}

func NewDocumentService() *DocumentService {
	that := &DocumentService{
		DefaultService:                    NewDefaultService(&models.Document{}),
		DocumentPermissionService:         NewDocumentPermissionService(),
		DocumentAccessRecordService:       NewDocumentAccessRecordService(),
		DocumentFavoritesService:          NewDocumentFavoritesService(),
		DocumentPermissionRequestsService: NewDocumentPermissionRequestsService(),
		DocumentVersionService:            NewDocumentVersionService(),
	}
	that.That = that
	return that
}

// type User struct {
// 	Id       int64  `json:"id"`
// 	Nickname string `json:"nickname"`
// 	Avatar   string `json:"avatar"`
// }

// func (user User) MarshalJSON() ([]byte, error) {
// 	// todo
// 	// if strings.HasPrefix(user.Avatar, "/") {
// 	// 	user.Avatar = config.Config.StorageUrl.Attatch + user.Avatar
// 	// }
// 	return models.MarshalJSON(user)
// }

// type Document struct {
// 	Id        int64            `json:"id"`
// 	CreatedAt time.Time        `json:"created_at"`
// 	DeletedAt models.DeletedAt `json:"deleted_at"`
// 	UserId    string           `json:"user_id"`
// 	Path      string           `json:"path"`
// 	DocType   models.DocType   `json:"doc_type"`
// 	Name      string           `json:"name"`
// 	Size      uint64           `json:"size"`
// 	VersionId string           `json:"version_id"`
// 	TeamId    string           `json:"team_id"`
// 	ProjectId string           `json:"project_id"`
// 	DeleteBy  int64            `json:"delete_by"`
// 	LockedAt  time.Time        `json:"locked_at"`
// }

// func (model Document) MarshalJSON() ([]byte, error) {
// 	return models.MarshalJSON(model)
// }

// type DocumentTeam struct {
// 	Id   int64  `json:"id"`
// 	Name string `json:"name"`
// }

// func (model DocumentTeam) MarshalJSON() ([]byte, error) {
// 	return models.MarshalJSON(model)
// }

// type DocumentProject struct {
// 	Id   int64  `json:"id"`
// 	Name string `json:"name"`
// }

// func (model DocumentProject) MarshalJSON() ([]byte, error) {
// 	return models.MarshalJSON(model)
// }

// type DocumentFavorites struct {
// 	Id         int64 `json:"id"`
// 	IsFavorite bool  `json:"is_favorite"`
// }

// func (model DocumentFavorites) MarshalJSON() ([]byte, error) {
// 	return models.MarshalJSON(model)
// }

// type DocumentAccessRecord struct {
// 	Id             int64     `json:"id"`
// 	LastAccessTime time.Time `json:"last_access_time"`
// }

// func (model DocumentAccessRecord) MarshalJSON() ([]byte, error) {
// 	return models.MarshalJSON(model)
// }

// type DocumentPermission struct {
// 	Id             int64                 `json:"id"`
// 	PermType       models.PermType       `json:"perm_type"`
// 	PermSourceType models.PermSourceType `json:"perm_source_type"`
// }

// func (model DocumentPermission) MarshalJSON() ([]byte, error) {
// 	return models.MarshalJSON(model)
// }

type DocumentPermissionRequests models.DocumentPermissionRequests

func (model DocumentPermissionRequests) MarshalJSON() ([]byte, error) {
	return models.MarshalJSON(model)
}

// 这得联合好几个表，唉
type DocumentQueryResItem struct {
	Document       models.Document     `gorm:"embedded;embeddedPrefix:document__" json:"document" join:";inner;id,[#document_id document_id]"`
	User           *models.UserProfile `gorm:"-" json:"user"`
	Team           *models.Team        `gorm:"embedded;embeddedPrefix:team__" json:"team" join:";left;id,document.team_id"`
	Project        *models.Project     `gorm:"embedded;embeddedPrefix:project__" json:"project" join:";left;id,document.project_id"`
	UserTeamMember *models.TeamMember  `gorm:"embedded;embeddedPrefix:tm__" json:"-" join:"team_member,tm;left;team_id,document.team_id;user_id,document.user_id;deleted_at,##is null"`
	// UserTeamNickname string             `gorm:"-" json:"user_team_nickname"`
}

type AccessRecordAndFavoritesQueryResItem struct {
	DocumentQueryResItem
	DocumentFavorites    models.DocumentFavorites    `gorm:"embedded;embeddedPrefix:document_favorites__" json:"document_favorites" join:";left;document_id,document.id;user_id,?user_id"`
	DocumentAccessRecord models.DocumentAccessRecord `gorm:"embedded;embeddedPrefix:document_access_record__" json:"document_access_record" join:";left;user_id,?user_id;document_id,document.id"`
}

type RecycleBinQueryResItem struct {
	AccessRecordAndFavoritesQueryResItem
	DeleteUser *models.UserProfile `gorm:"-" json:"delete_user"`
}

// FindRecycleBinByUserId 查询用户的回收站列表
func (s *DocumentService) FindRecycleBinByUserId(userId string, projectId string) *[]RecycleBinQueryResItem {
	var result = make([]RecycleBinQueryResItem, 0)
	whereArgsList := []WhereArgs{
		{"document.deleted_at is not null", nil},
	}
	if projectId != "" {
		whereArgsList = append(whereArgsList, WhereArgs{"document.project_id = ?", []any{projectId}})
	} else {
		whereArgsList = append(whereArgsList, WhereArgs{"document.user_id = ? and (document.project_id is null or document.project_id = '')", []any{userId}})
	}
	_ = s.Find(
		&result,
		&ParamArgs{"?user_id": userId},
		whereArgsList,
		&OrderLimitArgs{"document_access_record.last_access_time desc", 0},
		&Unscoped{},
	)
	// for i, _ := range result {
	// 	if result[i].UserTeamMember != nil {
	// 		result[i].UserTeamNickname = result[i].UserTeamMember.Nickname
	// 	}
	// }
	return &result
}

// FindDocumentByUserId 查询用户的文档列表
func (s *DocumentService) FindDocumentByUserId(userId string) *[]AccessRecordAndFavoritesQueryResItem {
	var result []AccessRecordAndFavoritesQueryResItem
	_ = s.Find(
		&result,
		&ParamArgs{"?user_id": userId},
		&WhereArgs{"document.user_id = ? and (document.project_id is null or document.project_id = '')", []any{userId}},
		&OrderLimitArgs{"document_access_record.last_access_time desc", 0},
	)
	return &result
}

// FindDocumentByUserIdWithCursor 使用游标分页查询用户的文档列表
func (s *DocumentService) FindDocumentByUserIdWithCursor(userId string, cursor string, limit int) (*[]AccessRecordAndFavoritesQueryResItem, bool) {
	var result []AccessRecordAndFavoritesQueryResItem

	// 基础查询条件
	whereArgs := WhereArgs{"document.user_id = ? and (document.project_id is null or document.project_id = '')", []any{userId}}

	// 如果有游标，添加游标条件
	if cursor != "" {
		// 解析游标（时间格式字符串）
		cursorTime, err := time.Parse(time.RFC3339, cursor)
		if err == nil {
			whereArgs = WhereArgs{
				"document.user_id = ? and (document.project_id is null or document.project_id = '') and document_access_record.last_access_time < ?",
				[]any{userId, cursorTime},
			}
		}
	}

	// 添加限制数量 +1，用于判断是否还有更多数据
	orderLimit := &OrderLimitArgs{"document_access_record.last_access_time desc", limit + 1}

	_ = s.Find(
		&result,
		&ParamArgs{"?user_id": userId},
		&whereArgs,
		orderLimit,
	)

	// 判断是否有更多数据
	hasMore := false
	if len(result) > limit {
		hasMore = true
		result = result[:limit] // 截取限制数量的数据
	}

	return &result, hasMore
}

// FindDocumentByProjectId 查询项目的文档列表
func (s *DocumentService) FindDocumentByProjectId(projectId string, userId string) *[]AccessRecordAndFavoritesQueryResItem {
	var result []AccessRecordAndFavoritesQueryResItem
	_ = s.Find(
		&result,
		&ParamArgs{"?user_id": userId},
		&WhereArgs{"document.project_id = ?", []any{projectId}},
		&OrderLimitArgs{"document_access_record.last_access_time desc", 0},
	)
	// for i, _ := range result {
	// 	if result[i].UserTeamMember != nil {
	// 		result[i].UserTeamNickname = result[i].UserTeamMember.Nickname
	// 	}
	// }
	return &result
}

// FindDocumentByProjectIdWithCursor 使用游标分页查询项目的文档列表
func (s *DocumentService) FindDocumentByProjectIdWithCursor(projectId string, userId string, cursor string, limit int) (*[]AccessRecordAndFavoritesQueryResItem, bool) {
	var result []AccessRecordAndFavoritesQueryResItem

	// 基础查询条件
	whereArgs := WhereArgs{"document.project_id = ?", []any{projectId}}

	// 如果有游标，添加游标条件
	if cursor != "" {
		// 解析游标（时间格式字符串）
		cursorTime, err := time.Parse(time.RFC3339, cursor)
		if err == nil {
			whereArgs = WhereArgs{
				"document.project_id = ? and document_access_record.last_access_time < ?",
				[]any{projectId, cursorTime},
			}
		}
	}

	// 添加限制数量 +1，用于判断是否还有更多数据
	orderLimit := &OrderLimitArgs{"document_access_record.last_access_time desc", limit + 1}

	_ = s.Find(
		&result,
		&ParamArgs{"?user_id": userId},
		&whereArgs,
		orderLimit,
	)

	// 判断是否有更多数据
	hasMore := false
	if len(result) > limit {
		hasMore = true
		result = result[:limit] // 截取限制数量的数据
	}

	return &result, hasMore
}

// FindAccessRecordsByUserId 查询用户的访问记录
func (s *DocumentService) FindAccessRecordsByUserId(userId string) *[]AccessRecordAndFavoritesQueryResItem {
	var result []AccessRecordAndFavoritesQueryResItem // 当指针的容器用
	err := s.DocumentAccessRecordService.Find(
		&result,
		&ParamArgs{"?user_id": userId},
		&WhereArgs{Query: "document_access_record.user_id = ? and document.deleted_at is null", Args: []any{userId}},
		&OrderLimitArgs{"document_access_record.last_access_time desc", 0},
	)
	if nil != err {
		log.Panicln("find access record err", err)
		return nil
	}
	return &result
}

// FindFavoritesByUserId 查询用户的收藏列表
func (s *DocumentService) FindFavoritesByUserId(userId string, projectId string) *[]AccessRecordAndFavoritesQueryResItem {
	var result []AccessRecordAndFavoritesQueryResItem
	whereArgsList := []WhereArgs{
		{"document_favorites.user_id = ? and document_favorites.is_favorite = 1 and document.deleted_at is null", []any{userId}},
	}
	if projectId != "" {
		whereArgsList = append(whereArgsList, WhereArgs{"document.project_id = ?", []any{projectId}})
	}
	_ = s.DocumentFavoritesService.Find(
		&result,
		&ParamArgs{"?user_id": userId},
		whereArgsList,
		&OrderLimitArgs{"document_access_record.last_access_time desc", 0},
	)
	return &result
}

func (s *DefaultService) AddLocked(info *models.DocumentLock) error {
	return s.DBModule.DB.Create(info).Error
}

func (s *DefaultService) AddLockedArr(info []models.DocumentLock) error {
	// 如果没有记录需要添加，直接返回
	if len(info) == 0 {
		return nil
	}

	// 使用事务来确保批量添加的原子性
	tx := s.DBModule.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 批量添加记录
	if err := tx.Create(&info).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	return tx.Commit().Error
}

// get locked
func (s *DocumentService) GetLocked(documentId string) ([]models.DocumentLock, error) {
	var locked []models.DocumentLock
	if err := s.DBModule.DB.Where("document_id = ?", documentId).Find(&locked).Error; err != nil {
		return nil, err
	}
	return locked, nil
}

func (s *DocumentService) DeleteAllLocked(documentId string) error {
	return s.DBModule.DB.Where("document_id = ?", documentId).Delete(&models.DocumentLock{}).Error
}

// delete locked
func (s *DocumentService) DeleteLocked(info *models.DocumentLock) error {
	return s.DBModule.DB.Delete(info).Error
}

func (s *DocumentService) DeleteAllLockedExcept(documentId string, except []models.DocumentLock) error {
	// 如果没有需要排除的记录，直接删除所有记录
	if len(except) == 0 {
		return s.DeleteAllLocked(documentId)
	}

	// 获取需要排除的记录ID
	var exceptIds []int64
	for _, lock := range except {
		exceptIds = append(exceptIds, lock.Id)
	}

	// 删除除了exceptIds之外的所有记录
	return s.DBModule.DB.Where("document_id = ? AND id NOT IN ?", documentId, exceptIds).Delete(&models.DocumentLock{}).Error
}

type DocumentSharesAndFavoritesQueryRes struct {
	AccessRecordAndFavoritesQueryResItem
	DocumentPermission models.DocumentPermission `gorm:"embedded;embeddedPrefix:document_permission__" json:"document_permission" join:";left;resource_type,?resource_type;resource_id,?resource_id;grantee_type,?grantee_type;grantee_id,?user_id"`
}

// FindSharesByUserId 查询用户加入的文档分享列表
func (s *DocumentService) FindSharesByUserId(userId string) *[]DocumentSharesAndFavoritesQueryRes {
	var result []DocumentSharesAndFavoritesQueryRes
	_ = s.DocumentPermissionService.Find(
		&result,
		&ParamArgs{"#document_id": "resource_id", "?user_id": userId},
		&WhereArgs{
			"document_permission.resource_type = ?" +
				" and document_permission.grantee_id = ?" +
				" and document.deleted_at is null",
			[]any{models.ResourceTypeDoc, userId},
		},
		&OrderLimitArgs{"document_access_record.last_access_time desc", 0},
	)
	return &result
}

type DocumentSharesQueryRes struct {
	DocumentQueryResItem
	DocumentPermission models.DocumentPermission `gorm:"embedded;embeddedPrefix:document_permission__" json:"document_permission" join:";left;resource_type,?resource_type;resource_id,?resource_id;grantee_type,?grantee_type;grantee_id,?user_id"`
}

// FindSharesByDocumentId 查询某个文档对所有用户的分享列表
func (s *DocumentService) FindSharesByDocumentId(documentId string) *[]DocumentSharesQueryRes {
	var result []DocumentSharesQueryRes
	_ = s.DocumentPermissionService.Find(
		&result,
		&ParamArgs{"#user_id": "document_permission.grantee_id"},
		&WhereArgs{
			"document_permission.resource_type = ?" +
				" and document_permission.resource_id = ?" +
				" and document.deleted_at is null",
			[]any{models.ResourceTypeDoc, documentId},
		},
		&ParamArgs{"#document_id": "resource_id"},
		&OrderLimitArgs{"document_permission.id desc", 0},
	)
	return &result
}

type DocumentInfoQueryRes struct {
	models.BaseModelStruct
	DocumentSharesAndFavoritesQueryRes
	DocumentPermissionRequests []DocumentPermissionRequests `gorm:"-" json:"document_permission_requests"`
	SharesCount                int64                        `gorm:"-" json:"shares_count"`
	ApplicationCount           int64                        `gorm:"-" json:"application_count"`
	// LockedInfo                 []models.DocumentLock        `gorm:"-" json:"locked_info"`
}

// GetDocumentInfoByDocumentAndUserId 查询某个文档对某个用户的信息
// 若传入permTypeForQueryApplicationCount不为models.PermTypeNone，则会返回该用户对该文档此权限的历史申请数量
func (s *DocumentService) GetDocumentInfoByDocumentAndUserId(documentId string, userId string, permTypeForQueryApplicationCount models.PermType) *DocumentInfoQueryRes {
	var result DocumentInfoQueryRes
	if err := s.Get(
		&result,
		"document.id = ?", documentId,
		&ParamArgs{"?user_id": userId, "?resource_type": models.ResourceTypeDoc, "?resource_id": documentId, "?grantee_type": models.GranteeTypeInternal},
		&OrderLimitArgs{"document_access_record.last_access_time desc", 0},
	); err != nil {
		return nil
	}

	if err := s.GetPermTypeByDocumentAndUserId(&result.DocumentPermission.PermType, documentId, userId); err != nil {
		return nil
	}

	_ = s.DocumentPermissionRequestsService.Find(&result.DocumentPermissionRequests, "user_id = ? and document_id = ?", userId, documentId)
	if permTypeForQueryApplicationCount != models.PermTypeNone {
		result.ApplicationCount = int64(len(sliceutil.FilterT(func(item DocumentPermissionRequests) bool {
			return item.PermType == permTypeForQueryApplicationCount
		}, result.DocumentPermissionRequests...)))
	}
	return &result
}

/*
文档权限优先级
1. 若为公共文档，取Max（现有权限，公共权限，项目权限）
2. 若为私有文档
	2.1 若不为项目文档，取现有权限
	2.2 若为项目文档
		2.2.1 若项目权限大于等于现有权限，取项目权限
		2.2.2 若项目权限小于现有权限（此时项目权限必为只读或可评论）
			2.2.2.1 若为创建者权限，取项目权限
			2.2.2.2 否则，取现有权限（此时为自定义权限）
其中，现有权限=创建者权限+已加入的自定义权限+已加入的公共权限
*/

// GetDocumentPermissionByDocumentAndUserId 获取用户的文档权限记录和用户的文档权限（包含文档本身的公共权限）
// 返回值第2个参数：是否为公共权限
func (s *DocumentService) GetDocumentPermissionByDocumentAndUserId(permType *models.PermType, documentId string, userId string) (*models.DocumentPermission, bool, error) {
	isPublicPerm := false

	var document models.Document
	if err := s.GetById(documentId, &document); err != nil {
		return nil, isPublicPerm, err
	}
	documentPermission := &models.DocumentPermission{}
	if err := s.DocumentPermissionService.Get(
		documentPermission,
		"resource_type = ? and resource_id = ? and grantee_id = ?",
		models.ResourceTypeDoc, documentId, userId,
	); err != nil && !errors.Is(err, ErrRecordNotFound) {
		return nil, isPublicPerm, err
	} else if errors.Is(err, ErrRecordNotFound) {
		documentPermission = nil
	}

	currentPermType := models.PermTypeNone // 现有权限
	publicPermType := models.PermTypeNone  // 公共权限
	projectPermType := models.PermTypeNone // 项目权限

	if documentPermission != nil {
		currentPermType = documentPermission.PermType
	}

	isPublic := document.DocType >= models.DocTypePublicReadable // 是否为公共文档
	if isPublic {
		if document.DocType == models.DocTypePublicReadable {
			publicPermType = models.PermTypeReadOnly
		} else if document.DocType == models.DocTypePublicCommentable {
			publicPermType = models.PermTypeCommentable
		} else if document.DocType == models.DocTypePublicEditable {
			publicPermType = models.PermTypeEditable
		}
	}

	isCreator := document.UserId == userId // 是否为创建者
	if isCreator {
		currentPermType = models.PermTypeEditable
	}

	isProjectDocument := document.ProjectId != "" // 是否为项目文档
	if isProjectDocument {
		projectService := NewProjectService()
		if _projectPermType, err := projectService.GetProjectPermTypeByForUser(document.ProjectId, userId); err == nil && _projectPermType != nil {
			projectPermType = (*_projectPermType).ToPermType()
		}
	}

	if isPublic { // 公共文档取Max（现有权限，公共权限，项目权限）
		*permType = models.PermType(math.Max(uint8(currentPermType), uint8(publicPermType), uint8(projectPermType)))
		if publicPermType > currentPermType && publicPermType > projectPermType {
			isPublicPerm = true
		}
	} else if !isProjectDocument { // 私有文档，非项目文档，取现有权限
		*permType = currentPermType
	} else if projectPermType >= currentPermType || isCreator { // 私有文档 && 项目文档， 项目权限大于等于现有权限，或项目权限小于现有权限且为创建者，取项目权限
		*permType = projectPermType
	} else { // 私有文档 && 项目文档，项目权限小于现有权限且不为创建者（即为自定义权限），取现有权限
		*permType = currentPermType
	}

	return documentPermission, isPublicPerm, nil
}

// GetPermTypeByDocumentAndUserId 获取用户对文档的权限（包含文档本身的公共权限）
func (s *DocumentService) GetPermTypeByDocumentAndUserId(permType *models.PermType, documentId string, userId string) error {
	if _, _, err := s.GetDocumentPermissionByDocumentAndUserId(permType, documentId, userId); err != nil {
		return err
	}
	return nil
}

type PermissionRequestsQueryResItem struct {
	DocumentQueryResItem
	DocumentPermissionRequests DocumentPermissionRequests `gorm:"embedded;embeddedPrefix:document_permission_requests__" json:"apply" table:""`
}

// FindPermissionRequests 获取用户所创建文档的权限申请列表
func (s *DocumentService) FindPermissionRequests(userId string, documentId string, startTime string) *[]PermissionRequestsQueryResItem {
	var result []PermissionRequestsQueryResItem
	whereArgsList := []WhereArgs{{Query: "document.user_id = ?", Args: []any{userId}}}
	if documentId != "" {
		whereArgsList = append(whereArgsList, WhereArgs{Query: "document.id = ?", Args: []any{documentId}})
	}
	if startTime != "" {
		whereArgsList = append(whereArgsList, WhereArgs{Query: "document_permission_requests.status = 0 and document_permission_requests.created_at >= ? and document_permission_requests.first_displayed_at is null", Args: []any{startTime}})
	}
	_ = s.DocumentPermissionRequestsService.Find(
		&result,
		&ParamArgs{"#user_id": "document_permission_requests.user_id"},
		whereArgsList,
		&OrderLimitArgs{"document_permission_requests.id desc", 0},
	)
	return &result
}

type DocumentPermissionService struct {
	*DefaultService
}

func NewDocumentPermissionService() *DocumentPermissionService {
	that := &DocumentPermissionService{
		DefaultService: NewDefaultService(&models.DocumentPermission{}),
	}
	that.That = that
	return that
}

type DocumentAccessRecordService struct {
	*DefaultService
}

func NewDocumentAccessRecordService() *DocumentAccessRecordService {
	that := &DocumentAccessRecordService{
		DefaultService: NewDefaultService(&models.DocumentAccessRecord{}),
	}
	that.That = that
	return that
}

type DocumentFavoritesService struct {
	*DefaultService
}

func NewDocumentFavoritesService() *DocumentFavoritesService {
	that := &DocumentFavoritesService{
		DefaultService: NewDefaultService(&models.DocumentFavorites{}),
	}
	that.That = that
	return that
}

type DocumentPermissionRequestsService struct {
	*DefaultService
}

func NewDocumentPermissionRequestsService() *DocumentPermissionRequestsService {
	that := &DocumentPermissionRequestsService{
		DefaultService: NewDefaultService(&models.DocumentPermissionRequests{}),
	}
	that.That = that
	return that
}

type DocumentVersionService struct {
	*DefaultService
}

func NewDocumentVersionService() *DocumentVersionService {
	that := &DocumentVersionService{
		DefaultService: NewDefaultService(&models.DocumentVersion{}),
	}
	that.That = that
	return that
}

type DocumentPermissionQuery struct {
	models.BaseModelStruct
	DocumentPermission models.DocumentPermission `gorm:"embedded;embeddedPrefix:document_permission__" json:"-" table:""`
	Document           models.Document           `gorm:"embedded;embeddedPrefix:document__" json:"-" join:";inner;id,resource_id"`
	// User               string                    `gorm:"embedded;embeddedPrefix:user__" json:"user" join:";inner;id,grantee_id"`
}

func (model DocumentPermissionQuery) MarshalJSON() ([]byte, error) {
	return models.MarshalJSON(model)
}

func (s *DocumentPermissionService) GetDocumentPermissionByPermId(documentPermissionId int64) (*DocumentPermissionQuery, error) {
	var result DocumentPermissionQuery
	if err := s.Get(
		&result,
		"document_permission.id = ?",
		documentPermissionId,
	); err != nil {
		return nil, err
	}
	return &result, nil
}
