package models

import (
	"protodesign.cn/kcserver/utils/time"
)

type ProjectPermType uint8

const (
	ProjectPermTypeNone        ProjectPermType = iota // 无权限
	ProjectPermTypeReadOnly                           // 只读
	ProjectPermTypeCommentable                        // 可评论
	ProjectPermTypeEditable                           // 可编辑
	ProjectPermTypeAdmin                              // 管理员
	ProjectPermTypeCreator                            // 创建者
)

// Project 项目
type Project struct {
	BaseModel
	TeamId        int64           `gorm:"not null" json:"team_id"`
	Name          string          `gorm:"size:64;not null" json:"name"`
	Description   string          `gorm:"size:128" json:"description"`
	PublicSwitch  bool            `gorm:"not null;default:false" json:"public_switch"`  // 是否在团队内部公开
	PermType      ProjectPermType `gorm:"default:1;not null" json:"perm_type"`          // 团队内的公开权限类型、或邀请权限类型
	InvitedSwitch bool            `gorm:"not null;default:false" json:"invited_switch"` // 邀请开关
	NeedApproval  bool            `gorm:"not null;default:true" json:"need_approval"`   // 申请是否需要审批
}

func (model Project) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

// ProjectPermSourceType 权限来源类型
type ProjectPermSourceType uint8

const (
	ProjectPermSourceTypeDefault ProjectPermSourceType = iota // 默认
	ProjectPermSourceTypeCustom                               // 自定义
)

// ProjectMember 项目成员
type ProjectMember struct {
	BaseModel
	ProjectId      int64                 `gorm:"not null" json:"project_id"`                 // 项目ID
	UserId         int64                 `gorm:"not null" json:"user_id"`                    // 用户ID
	PermType       ProjectPermType       `gorm:"default:1;not null" json:"perm_type"`        // 权限类型
	PermSourceType ProjectPermSourceType `gorm:"default:0;not null" json:"perm_source_type"` // 权限来源类型
}

func (model ProjectMember) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

// ProjectJoinRequestStatus 申请状态
type ProjectJoinRequestStatus uint8

const (
	ProjectJoinRequestStatusPending  ProjectJoinRequestStatus = iota // 等待审批
	ProjectJoinRequestStatusApproved                                 // 已批准
	ProjectJoinRequestStatusDenied                                   // 已拒绝
)

type ProjectJoinRequest struct {
	BaseModel
	UserId           int64                    `gorm:"index;not null" json:"user_id"`
	ProjectId        int64                    `gorm:"index;not null" json:"project_id"`
	PermType         ProjectPermType          `gorm:"not null" json:"perm_type"`
	Status           ProjectJoinRequestStatus `gorm:"not null;default:0" json:"status"`
	FirstDisplayedAt time.Time                `gorm:"" json:"first_displayed_at"`
	ProcessedAt      time.Time                `gorm:"" json:"processed_at"`
	ProcessedBy      int64                    `gorm:"" json:"processed_by"`
	ApplicantNotes   string                   `gorm:"size:256" json:"applicant_notes"`
	ProcessorNotes   string                   `gorm:"size:256" json:"processor_notes"`
}

func (model ProjectJoinRequest) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

// ProjectFavorite 项目收藏（固定）
type ProjectFavorite struct {
	BaseModel
	UserId    int64 `gorm:"uniqueIndex:idx_user_project;not null" json:"user_id"`
	ProjectId int64 `gorm:"uniqueIndex:idx_user_project;not null" json:"project_id"`
	IsFavor   bool  `gorm:"not null;default:true" json:"is_favor"`
}

func (model ProjectFavorite) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}
