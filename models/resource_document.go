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

type ResourceDocument struct {
	BaseModelStruct
	UserId      string `gorm:"uniqueIndex:idx_user_document,length:64" json:"user_id"`     // 用户ID
	DocumentId  string `gorm:"uniqueIndex:idx_user_document,length:32" json:"document_id"` // 文档ID
	References  int    `gorm:"" json:"references"`                                         // 引用次数
	Description string `gorm:"" json:"description"`                                        // 描述
}

func (model ResourceDocument) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model ResourceDocument) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

func (model ResourceDocument) GetId() interface{} {
	return model.Id
}

func (model ResourceDocument) TableName() string {
	return "resource_document"
}
