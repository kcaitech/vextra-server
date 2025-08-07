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

// DocumentFavorites 文档收藏
type DocumentFavorites struct {
	BaseModelStruct
	UserId     string `gorm:"uniqueIndex:idx_user_document,length:64" json:"user_id"`
	DocumentId string `gorm:"uniqueIndex:idx_user_document,length:32" json:"document_id"`
	IsFavorite bool   `gorm:"" json:"is_favorite"`
}

func (model DocumentFavorites) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model DocumentFavorites) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

func (model DocumentFavorites) GetId() interface{} {
	return model.Id
}

// tablename
func (model DocumentFavorites) TableName() string {
	return "document_favorites"
}
