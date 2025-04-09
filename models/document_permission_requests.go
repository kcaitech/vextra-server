package models

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/utils/time"
)

// StatusType 申请状态
type StatusType uint8

const (
	StatusTypePending  StatusType = iota // 等待审批
	StatusTypeApproved                   // 已批准
	StatusTypeDenied                     // 已拒绝
)

type DocumentPermissionRequests struct {
	BaseModelStruct
	UserId           string     `gorm:"index" json:"user_id"`
	DocumentId       string     `gorm:"index" json:"document_id"`
	PermType         PermType   `gorm:"" json:"perm_type"`
	Status           StatusType `gorm:"" json:"status"`
	FirstDisplayedAt time.Time  `gorm:"" json:"first_displayed_at"`
	ProcessedAt      time.Time  `gorm:"" json:"processed_at"`
	ProcessedBy      string     `gorm:"" json:"processed_by"`
	ApplicantNotes   string     `gorm:"size:256" json:"applicant_notes"`
	ProcessorNotes   string     `gorm:"size:256" json:"processor_notes"`
}

func (model DocumentPermissionRequests) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model DocumentPermissionRequests) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

func (model DocumentPermissionRequests) GetId() interface{} {
	return model.Id
}

// tablename
func (model DocumentPermissionRequests) TableName() string {
	return "document_permission_requests"
}
