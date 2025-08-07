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

import (
	"time"

	"gorm.io/gorm"
)

// DocumentAccessRecord 文档访问记录
type DocumentAccessRecord struct {
	BaseModelStruct
	UserId         string    `gorm:"uniqueIndex:idx_user_document,length:64" json:"user_id"`     // 用户ID
	DocumentId     string    `gorm:"uniqueIndex:idx_user_document,length:32" json:"document_id"` // 文档ID
	LastAccessTime time.Time `gorm:"autoCreateTime;type:datetime(6)" json:"last_access_time"`    // 上次访问时间
}

func (model DocumentAccessRecord) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model DocumentAccessRecord) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

func (model DocumentAccessRecord) GetId() interface{} {
	return model.Id
}

// tablename
func (model DocumentAccessRecord) TableName() string {
	return "document_access_record"
}
