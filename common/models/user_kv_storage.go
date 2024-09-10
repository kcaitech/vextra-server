package models

// UserKVStorage UserKVStorage表
type UserKVStorage struct {
	BaseModel
	UserId int64  `gorm:"index;uniqueIndex:idx_user_key" json:"user_id"`
	Key    string `gorm:"index;uniqueIndex:idx_user_key;not null" json:"key"`
	Value  string `gorm:"not null" json:"value"`
}

func (model UserKVStorage) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}
