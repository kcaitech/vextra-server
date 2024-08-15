package main

import (
	"database/sql/driver"
	"encoding/json"
	"gorm.io/gorm"
	myTime "protodesign.cn/kcserver/utils/time"
	"time"
)

type DeletedAt gorm.DeletedAt

func (n *DeletedAt) Scan(value interface{}) error {
	return (*gorm.DeletedAt)(n).Scan(value)
}

func (n DeletedAt) Value() (driver.Value, error) {
	return gorm.DeletedAt(n).Value()
}

func (n DeletedAt) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return json.Marshal(myTime.Time(n.Time))
	}
	return gorm.DeletedAt(n).MarshalJSON()
}

func (n *DeletedAt) UnmarshalJSON(b []byte) error {
	var t myTime.Time
	err := t.UnmarshalJSON(b)
	if err == nil {
		if newB, err := t.MarshalJSON(); err == nil {
			b = newB
		}
	}
	return (*gorm.DeletedAt)(n).UnmarshalJSON(b)
}

type BaseModel struct {
	Id        int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime;type:datetime(6)" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;type:datetime(6)" json:"updated_at"`
	DeletedAt DeletedAt `gorm:"index" json:"deleted_at"`
}
