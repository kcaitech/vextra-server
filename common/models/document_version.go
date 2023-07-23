package models

// DocumentVersion 文档版本
type DocumentVersion struct {
	BaseModel
	DocumentId int64  `gorm:"index" json:"document_id"`
	VersionId  string `gorm:"index;size:36" json:"version_id"`
	LastCmdId  int64  `gorm:"" json:"last_cmd_id"`
}

func (model *DocumentVersion) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}

//func (model *Document) UnmarshalJSON(data []byte) error {
//	return UnmarshalJSON(model, data)
//}
