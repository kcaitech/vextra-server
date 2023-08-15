package models

// InviteCode 邀请码
type InviteCode struct {
	BaseModel
	Code   string `gorm:"size:8;unique" json:"code"`
	UserId int64  `gorm:"" json:"user_id"`
}

func (model InviteCode) MarshalJSON() ([]byte, error) {
	return MarshalJSON(model)
}
