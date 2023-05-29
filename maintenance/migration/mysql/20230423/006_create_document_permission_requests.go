package main

import (
	"gorm.io/gorm"
	"protodesign.cn/kcserver/utils/time"
)

// StatusType 申请状态
type StatusType uint8

const (
	StatusTypePending  StatusType = iota // 等待审批
	StatusTypeApproved                   // 已批准
	StatusTypeDenied                     // 已拒绝
)

type DocumentPermissionRequests struct {
	BaseModel
	UserId           int64      `gorm:"index" json:"user_id"`
	DocumentId       int64      `gorm:"index" json:"document_id"`
	PermType         PermType   `gorm:"" json:"perm_type"`
	Status           StatusType `gorm:"" json:"status"`
	FirstDisplayedAt time.Time  `gorm:"" json:"first_displayed_at"`
	ProcessedAt      time.Time  `gorm:"" json:"processed_at"`
	ProcessedBy      int64      `gorm:"" json:"processed_by"`
	ApplicantNotes   string     `gorm:"size:256" json:"applicant_notes"`
	ProcessorNotes   string     `gorm:"size:256" json:"processor_notes"`
}

func DocumentPermissionRequestsUp(db *gorm.DB) error {
	if err := db.AutoMigrate(&DocumentPermissionRequests{}); err != nil {
		return err
	}
	return nil
}

func DocumentPermissionRequestsDown(db *gorm.DB) error {
	if err := db.Migrator().DropTable(&DocumentPermissionRequests{}); err != nil {
		return err
	}
	return nil
}
