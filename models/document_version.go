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

// DocumentVersion 文档版本
type DocumentVersion struct {
	BaseModelStruct
	DocumentId   string `gorm:"index" json:"document_id"`
	VersionId    string `gorm:"index;size:64" json:"version_id"` // 这是个oss的版本id
	LastCmdVerId uint   `gorm:"" json:"last_cmd_ver_id"`         // 此版本最后一个cmd的ver_id
}

func (model DocumentVersion) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

//func (model *Document) UnmarshalJSON(data []byte) error {
//	return UnmarshalJSON(model, data)
//}

func (model DocumentVersion) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model)
}

func (model DocumentVersion) GetId() interface{} {
	return model.Id
}

// tablename
func (model DocumentVersion) TableName() string {
	return "document_version"
}

type DocumentVersionWSData struct {
	DocumentId       string `json:"document_id"`
	VersionId        string `json:"version_id"`
	VersionStartWith uint   `json:"version_start_with"`
}
