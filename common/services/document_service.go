package services

import (
	"gorm.io/gorm"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/utils/time"
)

type DocumentService struct {
	*DefaultService
	DocumentPermissionService         *DocumentPermissionService
	DocumentAccessRecordService       *DocumentAccessRecordService
	DocumentFavoritesService          *DocumentFavoritesService
	DocumentPermissionRequestsService *DocumentPermissionRequestsService
}

func NewDocumentService() *DocumentService {
	that := &DocumentService{
		DefaultService:                    NewDefaultService(&models.Document{}),
		DocumentPermissionService:         NewDocumentPermissionService(),
		DocumentAccessRecordService:       NewDocumentAccessRecordService(),
		DocumentFavoritesService:          NewDocumentFavoritesService(),
		DocumentPermissionRequestsService: NewDocumentPermissionRequestsService(),
	}
	that.That = that
	return that
}

type DocumentQueryResItem struct {
	User struct {
		Id       int64  `json:"id"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	} `gorm:"embedded" json:"user" anonymous:"true"`
	Document struct {
		Id        int64          `json:"id"`
		CreatedAt time.Time      `json:"created_at"`
		DeletedAt gorm.DeletedAt `json:"deleted_at"`
		UserId    int64          `json:"user_id"`
		Path      string         `json:"path"`
		DocType   models.DocType `json:"doc_type"`
		Name      string         `json:"name"`
		Size      uint64         `json:"size"`
	} `gorm:"embedded" json:"document" anonymous:"true"`
	IsFavorite bool `json:"is_favorite"`
}

func (model *DocumentQueryResItem) MarshalJSON() ([]byte, error) {
	return models.MarshalJSON(model)
}

type AccessRecordQueryResItem struct {
	DocumentQueryResItem
	LastAccessTime time.Time `json:"last_access_time"`
}

// FindAccessRecordsByUserId 查询用户的访问记录
func (s *DocumentService) FindAccessRecordsByUserId(userId int64) *[]AccessRecordQueryResItem {
	var result []AccessRecordQueryResItem
	_ = s.DocumentAccessRecordService.Find(
		&result,
		WhereArgs{"document_access_record.user_id = ? and document.deleted_at is null", []any{userId}},
		JoinArgs{"inner join user on user.id = document_access_record.user_id", nil},
		JoinArgs{"inner join document on document.id = document_access_record.document_id", nil},
		JoinArgs{"left join document_favorites on document_favorites.user_id = document_access_record.user_id and document_favorites.document_id = document_access_record.document_id", nil},
		SelectArgs{"user.*, document.*, document_favorites.is_favorite, document_access_record.last_access_time", nil},
		OrderLimitArgs{"document_access_record.last_access_time desc", 0},
	)
	return &result
}

// FindFavoritesByUserId 查询用户的收藏列表
func (s *DocumentService) FindFavoritesByUserId(userId int64) *[]AccessRecordQueryResItem {
	var result []AccessRecordQueryResItem
	_ = s.DocumentFavoritesService.Find(
		&result,
		WhereArgs{"document_favorites.user_id = ? and document.deleted_at is null", []any{userId}},
		JoinArgs{"inner join user on user.id = document_favorites.user_id", nil},
		JoinArgs{"inner join document on document.id = document_favorites.document_id", nil},
		JoinArgs{"left join document_access_record on document_access_record.user_id = document_favorites.user_id and document_access_record.document_id = document_favorites.document_id", nil},
		SelectArgs{"user.*, document.*, document_favorites.is_favorite, document_access_record.last_access_time", nil},
		OrderLimitArgs{"document_access_record.last_access_time desc", 0},
	)
	return &result
}

type DocumentSharesQueryRes struct {
	DocumentQueryResItem
	LastAccessTime time.Time       `json:"last_access_time"`
	PermType       models.PermType `json:"perm_type"`
	Id             int64           `json:"id"`
}

func (model *DocumentSharesQueryRes) MarshalJSON() ([]byte, error) {
	return models.MarshalJSON(model)
}

// FindSharesByUserId 查询用户加入的文档分享列表
func (s *DocumentService) FindSharesByUserId(userId int64) *[]DocumentSharesQueryRes {
	var result []DocumentSharesQueryRes
	_ = s.DocumentPermissionService.Find(
		&result,
		WhereArgs{
			"document_permission.resource_type = ?" +
				" and document_permission.grantee_type = ?" +
				" and document_permission.grantee_id = ?" +
				" and document.deleted_at is null",
			[]any{models.ResourceTypeDoc, models.GranteeTypeExternal, userId},
		},
		JoinArgs{"inner join user on user.id = document_permission.grantee_id", nil},
		JoinArgs{"inner join document on document.id = document_permission.resource_id", nil},
		JoinArgs{"left join document_access_record on document_access_record.user_id = document_permission.grantee_id and document_access_record.document_id = document_permission.resource_id", nil},
		JoinArgs{"left join document_favorites on document_favorites.user_id = document_permission.grantee_id and document_favorites.document_id = document_permission.resource_id", nil},
		SelectArgs{"user.*, document.*, document_favorites.is_favorite, document_access_record.last_access_time, document_permission.perm_type, document_permission.id", nil},
		OrderLimitArgs{"document_access_record.last_access_time desc", 0},
	)
	return &result
}

// FindSharesByDocumentId 查询某个文档对所有用户的分享列表
func (s *DocumentService) FindSharesByDocumentId(documentId int64) *[]DocumentSharesQueryRes {
	var result []DocumentSharesQueryRes
	_ = s.DocumentPermissionService.Find(
		&result,
		WhereArgs{
			"document_permission.resource_type = ?" +
				" and document_permission.resource_id = ?" +
				" and document_permission.grantee_type = ?" +
				" and document.deleted_at is null",
			[]any{models.ResourceTypeDoc, documentId, models.GranteeTypeExternal},
		},
		JoinArgs{"inner join user on user.id = document_permission.grantee_id", nil},
		JoinArgs{"inner join document on document.id = document_permission.resource_id", nil},
		SelectArgs{"user.*, document.*, document_permission.perm_type, document_permission.id", nil},
		OrderLimitArgs{"document_permission.id desc", 0},
	)
	return &result
}

type DocumentInfoQueryRes struct {
	models.DefaultModelData
	DocumentQueryResItem
	PermType         models.PermType `json:"perm_type"`
	SharesCount      int64           `json:"shares_count"`
	ApplicationCount int64           `json:"application_count"`
}

// GetDocumentInfoByDocumentAndUserId 查询某个文档对某个用户的信息
func (s *DocumentService) GetDocumentInfoByDocumentAndUserId(documentId int64, userId int64, permType models.PermType) *DocumentInfoQueryRes {
	var result DocumentInfoQueryRes
	_ = s.Get(
		&result,
		"document.id = ?", documentId,
		JoinArgs{"inner join user on user.id = document.user_id", nil},
		JoinArgs{"left join document_access_record on document_access_record.document_id = document.id and document_access_record.user_id = ?", []any{userId}},
		JoinArgs{
			"left join document_permission on" +
				" document_permission.resource_type = ?" +
				" and document_permission.resource_id = ?" +
				" and document_permission.grantee_type = ?" +
				" and document_permission.grantee_id = ?",
			[]any{models.ResourceTypeDoc, documentId, models.GranteeTypeExternal, userId},
		},
		JoinArgs{"left join document_favorites on document_favorites.user_id = document.user_id and document_favorites.document_id = document.id", nil},
		SelectArgs{"user.*, document.*, document_favorites.is_favorite, document_permission.perm_type", nil},
		OrderLimitArgs{"document_access_record.last_access_time desc", 0},
	)
	if result.User.Id == userId {
		result.PermType = models.PermTypeEditable
	} else if result.Document.DocType == models.DocTypePrivate {
		result.PermType = models.PermTypeNone
	}
	if result.PermType == models.PermTypeNone {
		switch result.Document.DocType {
		case models.DocTypePublicReadable:
			result.PermType = models.PermTypeReadOnly
		case models.DocTypePublicCommentable:
			result.PermType = models.PermTypeCommentable
		case models.DocTypePublicEditable:
			result.PermType = models.PermTypeEditable
		}
	}
	_ = s.DocumentPermissionService.Count(
		&result.SharesCount,
		"resource_type = ? and resource_id = ? and grantee_type = ?",
		models.ResourceTypeDoc, documentId, models.GranteeTypeExternal,
	)
	whereQuery := "user_id = ? and document_id = ? and grantee_type = ?"
	whereArgs := []any{models.ResourceTypeDoc, documentId, models.GranteeTypeExternal}
	if permType != models.PermTypeNone {
		whereQuery += " and perm_type = ?"
		whereArgs = append(whereArgs, permType)
	}
	args := make([]any, len(whereArgs)+1)
	args[0] = whereQuery
	copy(args[1:], whereArgs)
	_ = s.DocumentPermissionRequestsService.Count(&result.ApplicationCount, args...)
	return &result
}

// GetSelfDocumentPermissionByDocumentAndUserId 获取用户对文档的权限（不包含文档本身的公共权限）
func (s *DocumentService) GetSelfDocumentPermissionByDocumentAndUserId(permType *models.PermType, documentId int64, userId int64) error {
	var document models.Document
	if err := s.GetById(documentId, &document); err != nil {
		return err
	}
	if document.UserId == userId {
		*permType = models.PermTypeEditable
		return nil
	}
	var documentPermission models.DocumentPermission
	if err := s.DocumentPermissionService.Get(
		&documentPermission,
		"resource_type = ? and resource_id = ? and grantee_type = ? and grantee_id = ?",
		models.ResourceTypeDoc, documentId, models.GranteeTypeExternal, userId,
	); err != nil && err != ErrRecordNotFound {
		return err
	}
	*permType = documentPermission.PermType
	return nil
}

// GetDocumentPermissionByDocumentAndUserId 获取用户对文档的权限（包含文档本身的公共权限）
func (s *DocumentService) GetDocumentPermissionByDocumentAndUserId(permType *models.PermType, documentId int64, userId int64) error {
	var document models.Document
	if err := s.GetById(documentId, &document); err != nil {
		return err
	}
	if document.UserId == userId {
		*permType = models.PermTypeEditable
		return nil
	}
	var documentPermission models.DocumentPermission
	if err := s.DocumentPermissionService.Get(
		&documentPermission,
		"resource_type = ? and resource_id = ? and grantee_type = ? and grantee_id = ?",
		models.ResourceTypeDoc, documentId, models.GranteeTypeExternal, userId,
	); err != nil && err != ErrRecordNotFound {
		return err
	}
	*permType = documentPermission.PermType
	if *permType == models.PermTypeNone {
		switch document.DocType {
		case models.DocTypePublicReadable:
			*permType = models.PermTypeReadOnly
		case models.DocTypePublicCommentable:
			*permType = models.PermTypeCommentable
		case models.DocTypePublicEditable:
			*permType = models.PermTypeEditable
		}
	}
	return nil
}

type PermissionRequestsQueryResItem struct {
	DocumentQueryResItem
	DocumentPermissionRequests models.DocumentPermissionRequests `gorm:"embedded" json:"apply"`
	PermType                   models.PermType                   `json:"perm_type"`
}

func (model *PermissionRequestsQueryResItem) MarshalJSON() ([]byte, error) {
	return models.MarshalJSON(model)
}

// FindPermissionRequests 获取用户创建的文档的权限申请列表
func (s *DocumentService) FindPermissionRequests(userId int64, documentId int64, startTime string) *[]PermissionRequestsQueryResItem {
	var result []PermissionRequestsQueryResItem
	whereArgs := WhereArgs{Query: "document.user_id = ?", Args: []any{userId}}
	if documentId != 0 {
		whereArgs.Query += " and document.id = ?"
		whereArgs.Args = append(whereArgs.Args, documentId)
	}
	if startTime != "" {
		whereArgs.Query += " and document_permission_requests.created_at >= ? and document_permission_requests.first_displayed_at is null"
		whereArgs.Args = append(whereArgs.Args, startTime)
	}
	_ = s.DocumentPermissionRequestsService.Find(
		&result,
		whereArgs,
		JoinArgs{"inner join user on user.id = document_permission_requests.user_id", nil},
		JoinArgs{"inner join document on document.id = document_permission_requests.document_id", nil},
		SelectArgs{"user.*, document.*, document_permission_requests.*, document_permission.perm_type", nil},
		OrderLimitArgs{"document_permission_requests.id desc", 0},
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
