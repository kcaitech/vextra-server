package models

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/utils/time"
)

type TeamPermType uint8

const (
	TeamPermTypeReadOnly TeamPermType = iota // 只读
	TeamPermTypeEditable                     // 可编辑
	TeamPermTypeAdmin                        // 管理员
	TeamPermTypeCreator                      // 创建者
	TeamPermTypeNone     TeamPermType = 255  // 无权限
)

// Team 团队
type Team struct {
	BaseModelStruct
	Name        string `gorm:"size:64;not null" json:"name"`
	Description string `gorm:"size:128" json:"description"`
	Avatar      string `gorm:"size:256" json:"avatar"`
	// Uid             string       `gorm:"unique;size:64" json:"uid"`                   // 团队ID?
	InvitedPermType TeamPermType `gorm:"not null;default:0" json:"invited_perm_type"` // 邀请权限类型
	OpenInvite      bool         `gorm:"not null;default:false" json:"open_invite"`   // 是否可以邀请
}

func (model Team) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model Team) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

func (model Team) GetId() int64 {
	return model.Id
}

// tablename
func (model Team) TableName() string {
	return "teams"
}

// TeamMember 团队成员
type TeamMember struct {
	BaseModelStruct
	TeamId   int64        `gorm:"not null" json:"team_id"`          // 团队ID
	UserId   string       `gorm:"not null" json:"user_id"`          // 用户ID
	PermType TeamPermType `gorm:"not null" json:"perm_type"`        // 权限类型
	Nickname string       `gorm:"size:64;not null" json:"nickname"` // 团队成员昵称
}

func (model TeamMember) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model TeamMember) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

func (model TeamMember) GetId() int64 {
	return model.Id
}

// tablename
func (model TeamMember) TableName() string {
	return "team_members"
}

// TeamJoinRequestStatus 申请状态
type TeamJoinRequestStatus uint8

const (
	TeamJoinRequestStatusPending  TeamJoinRequestStatus = iota // 等待审批
	TeamJoinRequestStatusApproved                              // 已批准
	TeamJoinRequestStatusDenied                                // 已拒绝
)

type TeamJoinRequest struct {
	BaseModelStruct
	UserId           string                `gorm:"index;not null" json:"user_id"`
	TeamId           int64                 `gorm:"index;not null" json:"team_id"`
	PermType         TeamPermType          `gorm:"not null" json:"perm_type"` // 取值：只读、可编辑
	Status           TeamJoinRequestStatus `gorm:"not null;default:0" json:"status"`
	FirstDisplayedAt time.Time             `gorm:"" json:"first_displayed_at"`
	ProcessedAt      time.Time             `gorm:"" json:"processed_at"`
	ProcessedBy      string                `gorm:"" json:"processed_by"`
	ApplicantNotes   string                `gorm:"size:256" json:"applicant_notes"`
	ProcessorNotes   string                `gorm:"size:256" json:"processor_notes"`
}

func (model TeamJoinRequest) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model TeamJoinRequest) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

func (model TeamJoinRequest) GetId() int64 {
	return model.Id
}

// tablename
func (model TeamJoinRequest) TableName() string {
	return "team_join_requests"
}

type TeamJoinRequestMessageShow struct {
	BaseModelStruct
	TeamJoinRequestId int64     `json:"team_join_request_id"`
	UserId            string    `json:"user_id"`
	TeamId            int64     `json:"team_id"`
	FirstDisplayedAt  time.Time `json:"first_displayed_at"`
}

func (model TeamJoinRequestMessageShow) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model TeamJoinRequestMessageShow) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

func (model TeamJoinRequestMessageShow) GetId() int64 {
	return model.Id
}

// tablename
func (model TeamJoinRequestMessageShow) TableName() string {
	return "team_join_request_message_shows"
}
