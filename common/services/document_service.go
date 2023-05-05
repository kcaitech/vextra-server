package services

import (
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/utils/time"
)

type DocumentService struct {
	DefaultService
	DocumentPermissionService   *DocumentPermissionService
	DocumentAccessRecordService *DocumentAccessRecordService
}

func NewDocumentService() *DocumentService {
	that := &DocumentService{
		DefaultService: DefaultService{
			DB:    models.DB,
			Model: &models.Document{},
		},
		DocumentPermissionService:   NewDocumentPermissionService(),
		DocumentAccessRecordService: NewDocumentAccessRecordService(),
	}
	that.That = that
	return that
}

type AccessRecordResult struct {
	models.Document
	LastAccessTime time.Time `json:"last_access_time"`
}

func (s *DocumentService) FindAccessRecordsByUserId(userId int64) *[]AccessRecordResult {
	var result []AccessRecordResult
	_ = s.DocumentAccessRecordService.Find(&result, "document_access_record.user_id = ?", userId,
		JoinArgs{"inner join user ON document_access_record.user_id = User.id", nil},
		JoinArgs{"inner join document ON document_access_record.document_id = Document.id", nil},
		SelectArgs{"document_access_record.*, document.*", nil},
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
