package models

import "protodesign.cn/kcserver/utils/time"

type ProjectType uint8

const (
	ProjectTypePrivate ProjectType = iota // 非公开
	ProjectTypePublic                     // 公开
)

type ProjectPermType uint8

const (
	ProjectPermNone        PermType = iota // 无权限
	ProjectPermReadOnly                    // 只读
	ProjectPermCommentable                 // 可评论
	ProjectPermEditable                    // 可编辑
	ProjectPermAdmin                       // 管理员
)

// ProjectPermSourceType 权限来源类型
type ProjectPermSourceType uint8

const (
	ProjectPermSourceTypeDefault PermSourceType = iota // 默认
	ProjectPermSourceTypeCustom                        // 自定义
)

// Project 项目
type Project struct {
	BaseModel
	TeamId      int64           `gorm:"not null" json:"team_id"`
	Name        string          `gorm:"size:64;not null" json:"name"`
	Description string          `gorm:"size:128" json:"description"`
	ProjectType ProjectType     `gorm:"default:0;not null" json:"project_type"`
	PermType    ProjectPermType `gorm:"default:1;not null" json:"perm_type"`
}

func (model *Project) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

// ProjectMember 项目成员
type ProjectMember struct {
	BaseModel
	ProjectId      int64                 `gorm:"uniqueIndex:idx_project_member;not null" json:"project_id"` // 项目ID
	UserId         int64                 `gorm:"uniqueIndex:idx_project_member;not null" json:"user_id"`    // 用户ID
	PermType       ProjectPermType       `gorm:"default:1;not null" json:"perm_type"`                       // 权限类型
	PermSourceType ProjectPermSourceType `gorm:"default:0;not null" json:"perm_source_type"`                // 权限来源类型
}

func (model *ProjectMember) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

// ProjectPermissionRequestStatus 申请状态
type ProjectPermissionRequestStatus uint8

const (
	ProjectPermissionRequestStatusPending  ProjectPermissionRequestStatus = iota // 等待审批
	ProjectPermissionRequestStatusApproved                                       // 已批准
	ProjectPermissionRequestStatusDenied                                         // 已拒绝
)

type ProjectPermissionRequests struct {
	BaseModel
	UserId           int64                          `gorm:"index;not null" json:"user_id"`
	ProjectId        int64                          `gorm:"index;not null" json:"project_id"`
	PermType         ProjectPermType                `gorm:"not null" json:"perm_type"`
	Status           ProjectPermissionRequestStatus `gorm:"not null;default:0" json:"status"`
	FirstDisplayedAt time.Time                      `gorm:"" json:"first_displayed_at"`
	ProcessedAt      time.Time                      `gorm:"" json:"processed_at"`
	ProcessedBy      int64                          `gorm:"" json:"processed_by"`
	ApplicantNotes   string                         `gorm:"size:256" json:"applicant_notes"`
	ProcessorNotes   string                         `gorm:"size:256" json:"processor_notes"`
}

func (model *ProjectPermissionRequests) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}
