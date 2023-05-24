package str

import "github.com/google/uuid"

// GetUid 获取uuid
func GetUid() string {
	return uuid.NewString()
}
