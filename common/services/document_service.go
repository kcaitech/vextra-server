package services

import (
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/utils/time"
)

type DocumentService struct {
	DefaultService
	DocumentPermissionService         *DocumentPermissionService
	DocumentAccessRecordService       *DocumentAccessRecordService
	DocumentFavoritesService          *DocumentFavoritesService
	DocumentPermissionRequestsService *DocumentPermissionRequestsService
}

func NewDocumentService() *DocumentService {
	that := &DocumentService{
		DefaultService: DefaultService{
			DB:    models.DB,
			Model: &models.Document{},
		},
		DocumentPermissionService:         NewDocumentPermissionService(),
		DocumentAccessRecordService:       NewDocumentAccessRecordService(),
		DocumentFavoritesService:          NewDocumentFavoritesService(),
		DocumentPermissionRequestsService: NewDocumentPermissionRequestsService(),
	}
	that.That = that
	return that
}

type DocumentQueryResItem struct {
	Document models.Document `gorm:"embedded" json:"document"`
	User     models.User     `gorm:"embedded" json:"user"`
}

type AccessRecordQueryResItem struct {
	DocumentQueryResItem
	LastAccessTime time.Time `json:"last_access_time"`
}

// FindAccessRecordsByUserId 查询用户的访问记录
func (s *DocumentService) FindAccessRecordsByUserId(userId int64) *[]DocumentQueryResItem {
	var result []DocumentQueryResItem
	_ = s.DocumentAccessRecordService.Find(
		&result,
		WhereArgs{"document_access_record.user_id = ? and document.deleted_at is null", []interface{}{userId}},
		JoinArgs{"inner join user on user.id = document_access_record.user_id", nil},
		JoinArgs{"inner join document on document.id = document_access_record.document_id", nil},
		SelectArgs{"user.*, document.*, document_access_record.last_access_time", nil},
		OrderLimitArgs{"document_access_record.last_access_time desc", 0},
	)
	return &result
}

// FindFavoritesByUserId 查询用户的收藏列表
func (s *DocumentService) FindFavoritesByUserId(userId int64) *[]DocumentQueryResItem {
	var result []DocumentQueryResItem
	_ = s.DocumentFavoritesService.Find(
		&result,
		WhereArgs{"document_favorites.user_id = ? and document.deleted_at is null", []interface{}{userId}},
		JoinArgs{"inner join user on user.id = document_favorites.user_id", nil},
		JoinArgs{"inner join document on document.id = document_favorites.document_id", nil},
		JoinArgs{"left join document_access_record on document_access_record.user_id = document_favorites.user_id and document_access_record.document_id = document_favorites.document_id", nil},
		SelectArgs{"user.*, document.*, document_access_record.last_access_time", nil},
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
			[]interface{}{models.ResourceTypeDoc, models.GranteeTypeExternal, userId},
		},
		JoinArgs{"inner join user on user.id = document_permission.grantee_id", nil},
		JoinArgs{"inner join document on document.id = document_permission.resource_id", nil},
		JoinArgs{"left join document_access_record on document_access_record.user_id = document_permission.grantee_id and document_access_record.document_id = document_permission.resource_id", nil},
		SelectArgs{"user.*, document.*, document_access_record.last_access_time, document_permission.perm_type", nil},
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
			[]interface{}{models.ResourceTypeDoc, documentId, models.GranteeTypeExternal},
		},
		JoinArgs{"inner join user on user.id = document_permission.grantee_id", nil},
		JoinArgs{"inner join document on document.id = document_permission.resource_id", nil},
		SelectArgs{"user.*, document.*, document_permission.perm_type, document_permission.id as id", nil},
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
		JoinArgs{"left join document_access_record on document_access_record.document_id = document.id and document_access_record.user_id = ?", []interface{}{userId}},
		JoinArgs{
			"left join document_permission on" +
				" document_permission.resource_type = ?" +
				" and document_permission.resource_id = ?" +
				" and document_permission.grantee_type = ?" +
				" and document_permission.grantee_id = ?",
			[]interface{}{models.ResourceTypeDoc, documentId, models.GranteeTypeExternal, userId},
		},
		SelectArgs{"user.*, document.*, document_access_record.last_access_time, document_permission.perm_type", nil},
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
	whereArgs := []interface{}{models.ResourceTypeDoc, documentId, models.GranteeTypeExternal}
	if permType != models.PermTypeNone {
		whereQuery += " and perm_type = ?"
		whereArgs = append(whereArgs, permType)
	}
	args := make([]interface{}, len(whereArgs)+1)
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

// FindPermissionRequests 获取用户创建的文档的权限申请列表
func (s *DocumentService) FindPermissionRequests(userId int64, documentId int64, startTime string) *[]PermissionRequestsQueryResItem {
	var result []PermissionRequestsQueryResItem
	whereArgs := WhereArgs{Query: "document.user_id = ?", Args: []interface{}{userId}}
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
		SelectArgs{"user.*, document.*, document_permission.perm_type, document_permission_requests.*", nil},
		OrderLimitArgs{"document_permission_requests.id desc", 0},
	)
	return &result
}

type DocumentPermissionService struct {
	DefaultService
}

func NewDocumentPermissionService() *DocumentPermissionService {
	that := &DocumentPermissionService{
		DefaultService: DefaultService{
			DB:    models.DB,
			Model: &models.DocumentPermission{},
		},
	}
	that.That = that
	return that
}

type DocumentAccessRecordService struct {
	DefaultService
}

func NewDocumentAccessRecordService() *DocumentAccessRecordService {
	that := &DocumentAccessRecordService{
		DefaultService: DefaultService{
			DB:    models.DB,
			Model: &models.DocumentAccessRecord{},
		},
	}
	that.That = that
	return that
}

type DocumentFavoritesService struct {
	DefaultService
}

func NewDocumentFavoritesService() *DocumentFavoritesService {
	that := &DocumentFavoritesService{
		DefaultService: DefaultService{
			DB:    models.DB,
			Model: &models.DocumentFavorites{},
		},
	}
	that.That = that
	return that
}

type DocumentPermissionRequestsService struct {
	DefaultService
}

func NewDocumentPermissionRequestsService() *DocumentPermissionRequestsService {
	that := &DocumentPermissionRequestsService{
		DefaultService: DefaultService{
			DB:    models.DB,
			Model: &models.DocumentPermissionRequests{},
		},
	}
	that.That = that
	return that
}
