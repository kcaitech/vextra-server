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

type User struct {
	Id       int64  `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

func (model *User) MarshalJSON() ([]byte, error) {
	return models.MarshalJSON(model)
}

type Document struct {
	Id        int64          `json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
	UserId    int64          `json:"user_id"`
	Path      string         `json:"path"`
	DocType   models.DocType `json:"doc_type"`
	Name      string         `json:"name"`
	Size      uint64         `json:"size"`
}

func (model *Document) MarshalJSON() ([]byte, error) {
	return models.MarshalJSON(model)
}

type DocumentFavorites struct {
	Id         int64 `json:"id"`
	IsFavorite bool  `json:"is_favorite"`
}

func (model *DocumentFavorites) MarshalJSON() ([]byte, error) {
	return models.MarshalJSON(model)
}

type DocumentAccessRecord struct {
	Id             int64     `json:"id"`
	LastAccessTime time.Time `json:"last_access_time"`
}

func (model *DocumentAccessRecord) MarshalJSON() ([]byte, error) {
	return models.MarshalJSON(model)
}

type DocumentPermission struct {
	Id       int64           `json:"id"`
	PermType models.PermType `json:"perm_type"`
}

func (model *DocumentPermission) MarshalJSON() ([]byte, error) {
	return models.MarshalJSON(model)
}

type DocumentPermissionRequests models.DocumentPermissionRequests

func (model *DocumentPermissionRequests) MarshalJSON() ([]byte, error) {
	return models.MarshalJSON(model)
}

type DocumentQueryResItem struct {
	Document Document `gorm:"embedded;embeddedPrefix:document__" json:"document" table:"document" join:"document;inner;id,[#document_id document_id]"`
	User     User     `gorm:"embedded;embeddedPrefix:user__" json:"user" table:"user" join:"user;inner;id,[#user_id document.user_id]"`
}

type AccessRecordAndFavoritesQueryResItem struct {
	DocumentQueryResItem
	DocumentFavorites    DocumentFavorites    `gorm:"embedded;embeddedPrefix:document_favorites__" json:"document_favorites" table:"document_favorites" join:"document_favorites;left;document_id,document.id;user_id,?user_id"`
	DocumentAccessRecord DocumentAccessRecord `gorm:"embedded;embeddedPrefix:document_access_record__" json:"document_access_record" table:"document_access_record" join:"document_access_record;left;user_id,?user_id;document_id,document.id"`
}

// FindRecycleBinByUserId 查询用户的回收站列表
func (s *DocumentService) FindRecycleBinByUserId(userId int64) *[]AccessRecordAndFavoritesQueryResItem {
	var result []AccessRecordAndFavoritesQueryResItem
	_ = s.Find(
		&result,
		&ParamArgs{"?user_id": userId},
		&WhereArgs{"document.user_id = ? and document.deleted_at is not null and purged_at is null", []any{userId}},
		&OrderLimitArgs{"document_access_record.last_access_time desc", 0},
		&Unscoped{},
	)
	return &result
}

// FindDocumentByUserId 查询用户的文档列表
func (s *DocumentService) FindDocumentByUserId(userId int64) *[]AccessRecordAndFavoritesQueryResItem {
	var result []AccessRecordAndFavoritesQueryResItem
	_ = s.Find(
		&result,
		&ParamArgs{"?user_id": userId},
		&WhereArgs{"document.user_id = ?", []any{userId}},
		&OrderLimitArgs{"document_access_record.last_access_time desc", 0},
	)
	return &result
}

// FindAccessRecordsByUserId 查询用户的访问记录
func (s *DocumentService) FindAccessRecordsByUserId(userId int64) *[]AccessRecordAndFavoritesQueryResItem {
	var result []AccessRecordAndFavoritesQueryResItem
	_ = s.DocumentAccessRecordService.Find(
		&result,
		&ParamArgs{"?user_id": userId},
		&WhereArgs{Query: "document_access_record.user_id = ? and document.deleted_at is null", Args: []any{userId}},
		&OrderLimitArgs{"document_access_record.last_access_time desc", 0},
	)
	return &result
}

// FindFavoritesByUserId 查询用户的收藏列表
func (s *DocumentService) FindFavoritesByUserId(userId int64) *[]AccessRecordAndFavoritesQueryResItem {
	var result []AccessRecordAndFavoritesQueryResItem
	_ = s.DocumentFavoritesService.Find(
		&result,
		&ParamArgs{"?user_id": userId},
		&WhereArgs{"document_favorites.user_id = ? and document.deleted_at is null and document_favorites.is_favorite = 1", []any{userId}},
		&OrderLimitArgs{"document_access_record.last_access_time desc", 0},
	)
	return &result
}

type DocumentSharesAndFavoritesQueryRes struct {
	AccessRecordAndFavoritesQueryResItem
	DocumentPermission DocumentPermission `gorm:"embedded;embeddedPrefix:document_permission__" json:"document_permission" table:"document_permission" join:"document_permission;left;resource_type,?resource_type;resource_id,?resource_id;grantee_type,?grantee_type;grantee_id,?user_id"`
}

// FindSharesByUserId 查询用户加入的文档分享列表
func (s *DocumentService) FindSharesByUserId(userId int64) *[]DocumentSharesAndFavoritesQueryRes {
	var result []DocumentSharesAndFavoritesQueryRes
	_ = s.DocumentPermissionService.Find(
		&result,
		&ParamArgs{"#document_id": "resource_id", "?user_id": userId},
		&WhereArgs{
			"document_permission.resource_type = ?" +
				" and document_permission.grantee_type = ?" +
				" and document_permission.grantee_id = ?" +
				" and document.deleted_at is null",
			[]any{models.ResourceTypeDoc, models.GranteeTypeExternal, userId},
		},
		&OrderLimitArgs{"document_access_record.last_access_time desc", 0},
	)
	return &result
}

type DocumentSharesQueryRes struct {
	DocumentQueryResItem
	DocumentPermission DocumentPermission `gorm:"embedded;embeddedPrefix:document_permission__" json:"document_permission" table:"document_permission" join:"document_permission;left;resource_type,?resource_type;resource_id,?resource_id;grantee_type,?grantee_type;grantee_id,?user_id"`
}

// FindSharesByDocumentId 查询某个文档对所有用户的分享列表
func (s *DocumentService) FindSharesByDocumentId(documentId int64) *[]DocumentSharesQueryRes {
	var result []DocumentSharesQueryRes
	_ = s.DocumentPermissionService.Find(
		&result,
		&ParamArgs{"#user_id": "document_permission.grantee_id"},
		&WhereArgs{
			"document_permission.resource_type = ?" +
				" and document_permission.resource_id = ?" +
				" and document_permission.grantee_type = ?" +
				" and document.deleted_at is null",
			[]any{models.ResourceTypeDoc, documentId, models.GranteeTypeExternal},
		},
		&ParamArgs{"#document_id": "resource_id"},
		&OrderLimitArgs{"document_permission.id desc", 0},
	)
	return &result
}

type DocumentInfoQueryRes struct {
	models.DefaultModelData
	DocumentSharesAndFavoritesQueryRes
	SharesCount      int64 `json:"shares_count"`
	ApplicationCount int64 `json:"application_count"`
}

// GetDocumentInfoByDocumentAndUserId 查询某个文档对某个用户的信息
func (s *DocumentService) GetDocumentInfoByDocumentAndUserId(documentId int64, userId int64, permType models.PermType) *DocumentInfoQueryRes {
	var result DocumentInfoQueryRes
	if err := s.Get(
		&result,
		"document.id = ?", documentId,
		&ParamArgs{"?user_id": userId, "?resource_type": models.ResourceTypeDoc, "?resource_id": documentId, "?grantee_type": models.GranteeTypeExternal},
		&OrderLimitArgs{"document_access_record.last_access_time desc", 0},
	); err != nil {
		return nil
	}
	if result.User.Id == userId {
		result.DocumentPermission.PermType = models.PermTypeEditable
	} else if result.Document.DocType == models.DocTypePrivate {
		result.DocumentPermission.PermType = models.PermTypeNone
	}
	if result.DocumentPermission.PermType == models.PermTypeNone {
		switch result.Document.DocType {
		case models.DocTypePublicReadable:
			result.DocumentPermission.PermType = models.PermTypeReadOnly
		case models.DocTypePublicCommentable:
			result.DocumentPermission.PermType = models.PermTypeCommentable
		case models.DocTypePublicEditable:
			result.DocumentPermission.PermType = models.PermTypeEditable
		}
	}
	_ = s.DocumentPermissionService.Count(
		&result.SharesCount,
		"resource_type = ? and resource_id = ? and grantee_type = ?",
		models.ResourceTypeDoc, documentId, models.GranteeTypeExternal,
	)
	var whereArgsList []*WhereArgs
	whereArgsList = append(whereArgsList, &WhereArgs{
		"user_id = ? and document_id = ?",
		[]any{userId, documentId},
	})
	if permType != models.PermTypeNone {
		whereArgsList = append(whereArgsList, &WhereArgs{
			"perm_type = ?",
			[]any{permType},
		})
	}
	_ = s.DocumentPermissionRequestsService.Count(&result.ApplicationCount, whereArgsList)
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
	if document.DocType == models.DocTypePrivate {
		*permType = models.PermTypeNone
	} else if *permType == models.PermTypeNone {
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
	DocumentPermissionRequests DocumentPermissionRequests `gorm:"embedded;embeddedPrefix:document_permission_requests__" json:"apply" table:"document_permission_requests"`
}

// FindPermissionRequests 获取用户所创建文档的权限申请列表
func (s *DocumentService) FindPermissionRequests(userId int64, documentId int64, startTime string) *[]PermissionRequestsQueryResItem {
	var result []PermissionRequestsQueryResItem
	whereArgsList := []WhereArgs{{Query: "document.user_id = ?", Args: []any{userId}}}
	if documentId != 0 {
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
