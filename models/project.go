package models

import (
	"gorm.io/gorm"
	"kcaitech.com/kcserver/utils/time"
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

func (projectPermType ProjectPermType) ToPermType() PermType {
	permType := PermType(projectPermType)
	if permType < PermTypeNone {
		permType = PermTypeNone
	} else if permType > PermTypeEditable {
		permType = PermTypeEditable
	}
	return permType
}

// Project 项目
type Project struct {
	// BaseModelStruct
	Id        string    `gorm:"primaryKey" json:"id,string"` // 主键，自增
	CreatedAt time.Time `gorm:"autoCreateTime;type:datetime(6)" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;type:datetime(6)" json:"updated_at"`
	DeletedAt DeletedAt `gorm:"index" json:"deleted_at"`

	TeamId       string          `gorm:"not null" json:"team_id"`
	Name         string          `gorm:"size:64;not null" json:"name"`
	Description  string          `gorm:"size:128" json:"description"`
	IsPublic     bool            `gorm:"not null;default:false" json:"is_public"`    // 是否在团队内部公开
	PermType     ProjectPermType `gorm:"default:1;not null" json:"perm_type"`        // 团队内的公开权限类型、或邀请权限类型
	OpenInvite   bool            `gorm:"not null;default:false" json:"open_invite"`  // 是否可以邀请
	NeedApproval bool            `gorm:"not null;default:true" json:"need_approval"` // 申请是否需要审批
}

func (model Project) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model Project) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

func (model Project) GetId() interface{} {
	return model.Id
}

// tablename
func (model Project) TableName() string {
	return "project"
}

// ProjectPermSourceType 权限来源类型
type ProjectPermSourceType uint8

const (
	ProjectPermSourceTypeDefault ProjectPermSourceType = iota // 默认
	ProjectPermSourceTypeCustom                               // 自定义
)

// ProjectMember 项目成员
type ProjectMember struct {
	BaseModelStruct
	ProjectId      string                `gorm:"not null" json:"project_id"`                 // 项目ID
	UserId         string                `gorm:"not null" json:"user_id"`                    // 用户ID
	PermType       ProjectPermType       `gorm:"default:1;not null" json:"perm_type"`        // 权限类型
	PermSourceType ProjectPermSourceType `gorm:"default:0;not null" json:"perm_source_type"` // 权限来源类型
}

func (model ProjectMember) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model ProjectMember) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

func (model ProjectMember) GetId() interface{} {
	return model.Id
}

// tablename
func (model ProjectMember) TableName() string {
	return "project_member"
}

// ProjectJoinRequestStatus 申请状态
type ProjectJoinRequestStatus uint8

const (
	ProjectJoinRequestStatusPending  ProjectJoinRequestStatus = iota // 等待审批
	ProjectJoinRequestStatusApproved                                 // 已批准
	ProjectJoinRequestStatusDenied                                   // 已拒绝
)

type ProjectJoinRequest struct {
	BaseModelStruct
	UserId           string                   `gorm:"index;not null" json:"user_id"`
	ProjectId        string                   `gorm:"index;not null" json:"project_id"`
	PermType         ProjectPermType          `gorm:"not null" json:"perm_type"`
	Status           ProjectJoinRequestStatus `gorm:"not null;default:0" json:"status"`
	FirstDisplayedAt time.Time                `gorm:"" json:"first_displayed_at"`
	ProcessedAt      time.Time                `gorm:"" json:"processed_at"`
	ProcessedBy      string                   `gorm:"" json:"processed_by"`
	ApplicantNotes   string                   `gorm:"size:256" json:"applicant_notes"`
	ProcessorNotes   string                   `gorm:"size:256" json:"processor_notes"`
}

func (model ProjectJoinRequest) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model ProjectJoinRequest) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

func (model ProjectJoinRequest) GetId() interface{} {
	return model.Id
}

// tablename
func (model ProjectJoinRequest) TableName() string {
	return "project_join_request"
}

type ProjectJoinRequestMessageShow struct {
	BaseModelStruct
	ProjectJoinRequestId int64     `json:"project_join_request_id"`
	UserId               string    `json:"user_id"`
	ProjectId            string    `json:"project_id"`
	FirstDisplayedAt     time.Time `json:"first_displayed_at"`
}

func (model ProjectJoinRequestMessageShow) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model ProjectJoinRequestMessageShow) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

func (model ProjectJoinRequestMessageShow) GetId() interface{} {
	return model.Id
}

// tablename
func (model ProjectJoinRequestMessageShow) TableName() string {
	return "project_join_request_message_show"
}

// ProjectFavorite 项目收藏（固定）
type ProjectFavorite struct {
	BaseModelStruct
	UserId    string `gorm:"uniqueIndex:idx_user_project,length:64;not null" json:"user_id"`
	ProjectId string `gorm:"uniqueIndex:idx_user_project,length:32;not null" json:"project_id"`
	IsFavor   bool   `gorm:"not null;default:true" json:"is_favor"`
}

func (model ProjectFavorite) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model ProjectFavorite) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

func (model ProjectFavorite) GetId() interface{} {
	return model.Id
}

// tablename
func (model ProjectFavorite) TableName() string {
	return "project_favorite"
}
