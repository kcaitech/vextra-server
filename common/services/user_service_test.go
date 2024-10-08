package services

import (
	"log"
	"testing"

	"kcaitech.com/kcserver/common/config"
	"kcaitech.com/kcserver/common/models"
	"kcaitech.com/kcserver/common/snowflake"
)

type userInfoResp struct {
	Id       int64  `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

func init() {
	conf := &config.BaseConfiguration{
		DB: struct {
			DSN string `yaml:"dsn" json:"dsn"`
		}{
			DSN: "admin:kcai1212@tcp(127.0.0.1:33306)/kcserver?charset=utf8&parseTime=True&loc=Local",
		},
	}
	snowflake.Init("../snowflake/config/config.yaml")
	models.Init(conf)
}

func TestGetById(t *testing.T) {
	userId := int64(45469806136135680)
	userService := NewUserService()
	userList := make([]userInfoResp, 0)
	if err := userService.Find(&userList, "id = ?", userId); err != nil {
		log.Println(err)
		return
	}
	log.Println(userList)
}
