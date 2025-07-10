package models

import "gorm.io/gorm"

type AccessAuthPriorityMask uint64

const (
	// 文档操作权限
	AccessAuthPriorityMaskRead    = 1 << 0
	AccessAuthPriorityMaskComment = 1 << 1
	AccessAuthPriorityMaskWrite   = 1 << 2
	AccessAuthPriorityMaskDelete  = 1 << 3
	AccessAuthPriorityMaskCreate  = 1 << 4

	// todo
	// 团队操作权限
	// 项目操作权限

	AccessAuthPriorityMaskAll = AccessAuthPriorityMaskRead | AccessAuthPriorityMaskComment | AccessAuthPriorityMaskWrite | AccessAuthPriorityMaskDelete | AccessAuthPriorityMaskCreate
)

type AccessAuthRangeMask uint64

const (
	AccessAuthResourceMaskDocument = 1 << 0
	AccessAuthResourceMaskUser     = 1 << 1
	AccessAuthResourceMaskProject  = 1 << 2
	AccessAuthResourceMaskTeam     = 1 << 3

	AccessAuthResourceMaskAll = AccessAuthResourceMaskDocument | AccessAuthResourceMaskUser | AccessAuthResourceMaskProject | AccessAuthResourceMaskTeam
)

type AccessAuth struct {
	BaseModelStruct
	UserId       string `json:"user_id"`
	Key          string `gorm:"size:255;uniqueIndex" json:"key"`
	Secret       string `gorm:"size:255" json:"secret"`
	PriorityMask uint64 `json:"priority_mask"`
	ResourceMask uint64 `json:"resource_mask"`
}

func (model AccessAuth) GetId() interface{} {
	return model.Id
}

func (model AccessAuth) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model AccessAuth) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

// tablename
func (model AccessAuth) TableName() string {
	return "access_auth"
}

func (m *AccessAuth) HasReadPriority() bool {
	return m.PriorityMask&AccessAuthPriorityMaskRead != 0
}
func (m *AccessAuth) HasCommentPriority() bool {
	return m.PriorityMask&AccessAuthPriorityMaskComment != 0
}
func (m *AccessAuth) HasWritePriority() bool {
	return m.PriorityMask&AccessAuthPriorityMaskWrite != 0
}
func (m *AccessAuth) HasDeletePriority() bool {
	return m.PriorityMask&AccessAuthPriorityMaskDelete != 0
}
func (m *AccessAuth) HasCreatePriority() bool {
	return m.PriorityMask&AccessAuthPriorityMaskCreate != 0
}

func (m *AccessAuth) HasPriority(priority AccessAuthPriorityMask) bool {
	return m.PriorityMask&uint64(priority) != 0
}

func (m *AccessAuth) HasUserAccessRange() bool {
	return m.ResourceMask&AccessAuthResourceMaskUser != 0
}
func (m *AccessAuth) HasDocumentAccessRange() bool {
	return m.ResourceMask&AccessAuthResourceMaskDocument != 0
}
func (m *AccessAuth) HasProjectAccessRange() bool {
	return m.ResourceMask&AccessAuthResourceMaskProject != 0
}
func (m *AccessAuth) HasTeamAccessRange() bool {
	return m.ResourceMask&AccessAuthResourceMaskTeam != 0
}

func (m *AccessAuth) HasAccessRange(rangeMask AccessAuthRangeMask) bool {
	return m.ResourceMask&uint64(rangeMask) != 0
}

type AccessAuthResourceType uint8

const (
	AccessAuthResourceTypeDocument AccessAuthResourceType = iota
	AccessAuthResourceTypeProject  AccessAuthResourceType = iota
	AccessAuthResourceTypeTeam     AccessAuthResourceType = iota
)

type AccessAuthResource struct {
	BaseModelStruct
	Key        string `json:"key" gorm:"uniqueIndex:idx_key_resource_id"`
	Type       uint8  `json:"type" gorm:"uniqueIndex:idx_key_resource_id"`
	ResourceId string `json:"resource_id" gorm:"uniqueIndex:idx_key_resource_id"` // 与key的组合是唯一的
}

func (model AccessAuthResource) GetId() int64 {
	return model.Id
}

func (model AccessAuthResource) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model AccessAuthResource) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

// tablename
func (model AccessAuthResource) TableName() string {
	return "access_auth_range"
}

func (m *AccessAuthResource) IsDocumentResource() bool {
	return m.Type == uint8(AccessAuthResourceTypeDocument)
}
func (m *AccessAuthResource) IsProjectResource() bool {
	return m.Type == uint8(AccessAuthResourceTypeProject)
}
func (m *AccessAuthResource) IsTeamResource() bool {
	return m.Type == uint8(AccessAuthResourceTypeTeam)
}
