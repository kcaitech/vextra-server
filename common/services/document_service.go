package services

import (
	"protodesign.cn/kcserver/common/models"
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

func (s *DocumentService) FindAccessRecordsByUserId(userId int64) {

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
