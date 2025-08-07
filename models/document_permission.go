/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package models

import "gorm.io/gorm"

// ResourceType 资源类型
type ResourceType uint8

const (
	ResourceTypeDoc    ResourceType = iota // 文档
	ResourceTypeFolder                     // 文件夹
)

// GranteeType 受让人类型
type GranteeType uint8

const (
	GranteeTypeInternal   GranteeType = iota // 内部人员
	GranteeTypeDepartment                    // 部门
	GranteeTypeExternal                      // 外部人员
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
	BaseModelStruct
	ResourceType   ResourceType   `gorm:"default:0; uniqueIndex:unique_index" json:"resource_type"` // 资源类型，0-文档，1-文件夹
	ResourceId     string         `gorm:"uniqueIndex:unique_index,length:32" json:"resource_id"`    // 资源ID，文档ID或文件夹ID？
	GranteeType    GranteeType    `gorm:"default:0; uniqueIndex:unique_index" json:"grantee_type"`  // 受让人类型，0-外部人员，1-内部人员，2-部门
	GranteeId      string         `gorm:"uniqueIndex:unique_index,length:64" json:"grantee_id"`     // 受让人ID，用户id?
	PermType       PermType       `gorm:"default:1" json:"perm_type"`                               // 权限类型，0-无权限，1-只读，2-可评论，3-可编辑
	PermSourceType PermSourceType `gorm:"default:0" json:"perm_source_type"`                        // 权限来源类型，0-默认，1-自定义
}

func (model DocumentPermission) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model DocumentPermission) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

func (model DocumentPermission) GetId() interface{} {
	return model.Id
}

// tablename
func (model DocumentPermission) TableName() string {
	return "document_permission"
}
