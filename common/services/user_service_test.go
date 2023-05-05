package services

import (
	"log"
	"protodesign.cn/kcserver/common/config"
	"protodesign.cn/kcserver/common/models"
	"protodesign.cn/kcserver/common/snowflake"
	"testing"
)

type userInfoResp struct {
	Id       int64  `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

func init() {
	conf := &config.BaseConfiguration{
		DB: struct {
			DSN string `yaml:"dsn"`
		}{
			DSN: "root:_IpFCT^*pui~Mac7~0%SIicRq@Z6rtzE@tcp(localhost:33306)/kcserver?charset=utf8&parseTime=True&loc=Local",
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
