package models

// Options 配置表
type Options struct {
	BaseModel
	Type   string `gorm:"index;not null;size:255" json:"type"`
	Detail string `gorm:"not null" json:"detail"`
}

func (model Options) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}
