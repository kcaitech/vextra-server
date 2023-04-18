package models

import (
	"gorm.io/gorm"
	"time"
)

var DB *gorm.DB

type BaseModel struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
