package services

import (
	"protodesign.cn/kcserver/common/models"
)

type DocumentService struct {
	DefaultService
}

func NewDocumentService() *DocumentService {
	that := &DocumentService{
		DefaultService: DefaultService{
			DB:    models.DB,
			Model: &models.Document{},
		},
	}
	that.That = that
	return that
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
