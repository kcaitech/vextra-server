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

type DocumentUserService struct {
	DefaultService
}
