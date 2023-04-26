package main

import (
	"gorm.io/gorm"
)

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

type DocumentPermission struct {
	BaseModel
	ID           uint         `gorm:"primary_key" json:"id"`          // 授权记录的ID
	ResourceType ResourceType `gorm:"default:0" json:"resource_type"` // 资源类型
	ResourceId   uint         `gorm:"" json:"resourceId"`             // 资源ID
	GranteeType  GranteeType  `gorm:"default:0" json:"grantee_type"`  // 受让人类型
	GranteeId    uint         `gorm:"" json:"grantee_id"`             // 受让人ID
	PermType     PermType     `gorm:"default:0" json:"perm_type"`     // 权限类型
}

func DocumentPermissionUp(db *gorm.DB) error {
	if err := db.AutoMigrate(&DocumentPermission{}); err != nil {
		return err
	}
	return nil
}

func DocumentPermissionDown(db *gorm.DB) error {
	if err := db.Migrator().DropTable(&DocumentPermission{}); err != nil {
		return err
	}
	return nil
}
