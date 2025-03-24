package models

// ResourceType 资源类型
type ResourceType uint8

const (
	ResourceTypeDoc    ResourceType = iota // 文档
	ResourceTypeFolder                     // 文件夹
)

// GranteeType 受让人类型
type GranteeType uint8

const (
	GranteeTypeExternal   GranteeType = iota // 外部人员
	GranteeTypeInternal                      // 内部人员
	GranteeTypeDepartment                    // 部门
)

// PermType 权限类型
type PermType uint8

const (
	PermTypeNone        PermType = iota // 无权限
	PermTypeReadOnly                    // 只读
	PermTypeCommentable                 // 可评论
	PermTypeEditable                    // 可编辑
)

// PermSourceType 权限来源类型
type PermSourceType uint8

const (
	PermSourceTypeDefault PermSourceType = iota // 默认
	PermSourceTypeCustom                        // 自定义
)

// DocumentPermission 文档权限
type DocumentPermission struct {
	BaseModel
	ResourceType   ResourceType   `gorm:"default:0; uniqueIndex:unique_index" json:"resource_type"` // 资源类型
	ResourceId     int64          `gorm:"uniqueIndex:unique_index" json:"resource_id"`              // 资源ID
	GranteeType    GranteeType    `gorm:"default:0; uniqueIndex:unique_index" json:"grantee_type"`  // 受让人类型
	GranteeId      string         `gorm:"uniqueIndex:unique_index" json:"grantee_id"`               // 受让人ID
	PermType       PermType       `gorm:"default:1" json:"perm_type"`                               // 权限类型
	PermSourceType PermSourceType `gorm:"default:0" json:"perm_source_type"`                        // 权限来源类型
}

func (model DocumentPermission) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}
