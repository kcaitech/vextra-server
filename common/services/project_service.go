package services

import (
	"protodesign.cn/kcserver/common/models"
)

type ProjectService struct {
	*DefaultService
}

func NewProjectService() *ProjectService {
	that := &ProjectService{
		DefaultService: NewDefaultService(&models.Project{}),
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

type ProjectPermissionRequestsService struct {
	*DefaultService
}

func NewProjectPermissionRequestsService() *ProjectPermissionRequestsService {
	that := &ProjectPermissionRequestsService{
		DefaultService: NewDefaultService(&models.ProjectPermissionRequests{}),
	}
	that.That = that
	return that
}
