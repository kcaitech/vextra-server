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

// UserKVStorage UserKVStorage表
type UserKVStorage struct {
	BaseModelStruct
	UserId string `gorm:"not null;index;uniqueIndex:idx_user_key" json:"user_id"`
	Key    string `gorm:"not null;index;uniqueIndex:idx_user_key" json:"key"`
	Value  string `gorm:"not null" json:"value"`
}

func (model UserKVStorage) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

func (model UserKVStorage) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

func (model UserKVStorage) GetId() interface{} {
	return model.Id
}

// tablename
func (model UserKVStorage) TableName() string {
	return "user_kv_storage"
}
