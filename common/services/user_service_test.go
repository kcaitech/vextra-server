package services

import (
	"log"
	"testing"

	"kcaitech.com/kcserver/common/config"
	"kcaitech.com/kcserver/common/models"
	"kcaitech.com/kcserver/common/snowflake"
	snconf "kcaitech.com/kcserver/common/snowflake/config"
)

type userInfoResp struct {
	Id       int64  `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

func init() {
	conf := &config.BaseConfiguration{
		DB: struct {
			DSN string `yaml:"url" json:"url"`
		}{
			DSN: "admin:kcai1212@tcp(127.0.0.1:33306)/kcserver?charset=utf8&parseTime=True&loc=Local",
		},
	}

	snowflake.Init(&snconf.LoadConfig("../snowflake/config/config.yaml").Snowflake)
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
