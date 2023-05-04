package models

// ResourceType 资源类型
type ResourceType uint8 // 文档授权类型
const (
	ResourceTypeDoc    ResourceType = 0 // 文档
	ResourceTypeFolder ResourceType = 1 // 文件夹
)

// GranteeType 受让人类型
type GranteeType uint8 // 文档授权类型
const (
	GranteeTypeExternal   ResourceType = 0 // 外部人员
	GranteeTypeInternal   ResourceType = 1 // 内部人员
	GranteeTypeDepartment ResourceType = 2 // 部门
)

// PermType 权限类型
type PermType uint8 // 文档授权类型
const (
	PermTypeNone        PermType = 0 // 无权限
	PermTypeReadOnly    PermType = 1 // 只读
	PermTypeCommentable PermType = 2 // 可评论
	PermTypeEditable    PermType = 3 // 可编辑
)

// DocumentPermission 文档权限
type DocumentPermission struct {
	BaseModel
	ResourceType ResourceType `gorm:"default:0" json:"resource_type"` // 资源类型
	ResourceId   int64        `gorm:"" json:"resourceId"`             // 资源ID
	GranteeType  GranteeType  `gorm:"default:0" json:"grantee_type"`  // 受让人类型
	GranteeId    int64        `gorm:"" json:"grantee_id"`             // 受让人ID
	PermType     PermType     `gorm:"default:0" json:"perm_type"`     // 权限类型
}
