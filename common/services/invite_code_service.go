package services

import (
	"protodesign.cn/kcserver/common/models"
)

type InviteCodeService struct {
	*DefaultService
}

func NewInviteCodeService() *InviteCodeService {
	that := &InviteCodeService{
		DefaultService: NewDefaultService(&models.InviteCode{}),
	}
	that.That = that
	return that
}
