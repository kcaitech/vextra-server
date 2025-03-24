package models

import (
	"strings"

	"kcaitech.com/kcserver/common/config"
)

// 自己定义一份，防此auth端变更
type UserProfile struct {
	DefaultModelData
	Nickname string `json:"nickname"` // Nickname
	Avatar   string `json:"avatar"`   // Avatar URL
	UserId   string `json:"user_id"`  // User ID
}

func (user UserProfile) MarshalJSON() ([]byte, error) {
	if strings.HasPrefix(user.Avatar, "/") {
		user.Avatar = config.Config.StorageUrl.Attatch + user.Avatar
	}
	return MarshalJSON(user)
}
