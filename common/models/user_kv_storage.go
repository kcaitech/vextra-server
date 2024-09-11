package models

// UserKVStorage UserKVStorage表
type UserKVStorage struct {
	BaseModel
	UserId int64  `gorm:"not null;index;uniqueIndex:idx_user_key" json:"user_id"`
	Key    string `gorm:"not null;index;uniqueIndex:idx_user_key" json:"key"`
	Value  string `gorm:"not null" json:"value"`
}

func (model UserKVStorage) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}
