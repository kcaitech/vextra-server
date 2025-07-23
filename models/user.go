package models

// 自己定义一份，防此auth端变更
type UserProfile struct {
	// DefaultModelData
	Nickname string `json:"nickname"` // Nickname
	Avatar   string `json:"avatar"`   // Avatar URL
	Id       string `json:"id"`       // User ID
}

func (user UserProfile) MarshalJSON() ([]byte, error) {

	return MarshalJSON(user)
}
