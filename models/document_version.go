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
