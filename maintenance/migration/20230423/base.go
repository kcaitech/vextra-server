package main

import (
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	Id        int64          `gorm:"primaryKey;autoIncrement:false" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime;type:datetime(6)" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;type:datetime(6)" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
