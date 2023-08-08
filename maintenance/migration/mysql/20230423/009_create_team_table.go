package main

import (
	"gorm.io/gorm"
	"protodesign.cn/kcserver/utils/time"
)

const (
	TeamPermTypeReadOnly TeamPermType = iota // 只读
	TeamPermTypeEditable                     // 可编辑
	TeamPermTypeAdmin                        // 管理员
	TeamPermTypeCreator                      // 创建者
)

// Team 团队
type Team struct {
	BaseModel
	Name            string       `gorm:"size:64;not null" json:"name"`
	Description     string       `gorm:"size:128" json:"description"`
	Avatar          string       `gorm:"size:256" json:"avatar"`
	Uid             string       `gorm:"unique;size:64" json:"uid"`
	InvitedPermType TeamPermType `gorm:"not null" json:"invited_perm_type"`
}

type TeamPermType uint8

// TeamMember 团队成员
type TeamMember struct {
	BaseModel
	TeamId   int64        `gorm:"uniqueIndex:idx_team_member;not null" json:"team_id"` // 团队ID
	UserId   int64        `gorm:"uniqueIndex:idx_team_member;not null" json:"user_id"` // 用户ID
	PermType TeamPermType `gorm:"not null" json:"perm_type"`                           // 权限类型
}

// TeamJoinRequestStatus 申请状态
type TeamJoinRequestStatus uint8

const (
	TeamJoinRequestStatusPending  TeamJoinRequestStatus = iota // 等待审批
	TeamJoinRequestStatusApproved                              // 已批准
	TeamJoinRequestStatusDenied                                // 已拒绝
)

type TeamJoinRequest struct {
	BaseModel
	UserId           int64                 `gorm:"index;not null" json:"user_id"`
	TeamId           int64                 `gorm:"index;not null" json:"team_id"`
	PermType         TeamPermType          `gorm:"not null" json:"perm_type"` // 取值：可编辑、只读
	Status           TeamJoinRequestStatus `gorm:"not null;default:0" json:"status"`
	FirstDisplayedAt time.Time             `gorm:"" json:"first_displayed_at"`
	ProcessedAt      time.Time             `gorm:"" json:"processed_at"`
	ProcessedBy      int64                 `gorm:"" json:"processed_by"`
	ApplicantNotes   string                `gorm:"size:256" json:"applicant_notes"`
	ProcessorNotes   string                `gorm:"size:256" json:"processor_notes"`
}

func TeamUp(db *gorm.DB) error {
	if err := db.AutoMigrate(&Team{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&TeamMember{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&TeamJoinRequest{}); err != nil {
		return err
	}
	return nil
}

func TeamDown(db *gorm.DB) error {
	if err := db.Migrator().DropTable(&Team{}); err != nil {
		return err
	}
	if err := db.Migrator().DropTable(&TeamMember{}); err != nil {
		return err
	}
	if err := db.Migrator().DropTable(&TeamJoinRequest{}); err != nil {
		return err
	}
	return nil
}
