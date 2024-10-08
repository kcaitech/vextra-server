package controllers

import (
	"kcaitech.com/kcserver/common/config"
	jwt "kcaitech.com/kcserver/common/jwt/config"
	mongo "kcaitech.com/kcserver/common/mongo/config"
	redis "kcaitech.com/kcserver/common/redis/config"
	safereview "kcaitech.com/kcserver/common/safereview/config"
	snowflake "kcaitech.com/kcserver/common/snowflake/config"
	storage "kcaitech.com/kcserver/common/storage/config"
)

type Configuration struct {
	config.BaseConfiguration `yaml:",inline" json:",inline"`
	Wx                       struct {
		Appid  string `yaml:"appid" json:"appid"`
		Secret string `yaml:"secret" json:"secret"`
	} `yaml:"wx" json:"wx"`
	WxMp struct {
		Appid  string `yaml:"appid" json:"appid"`
		Secret string `yaml:"secret" json:"secret"`
	} `yaml:"wxMp" json:"wxMp"`
	DocumentVersionServer struct {
		Host string `yaml:"host" json:"host"`
	} `yaml:"documentVersionServer" json:"documentVersionServer"`

	Jwt       jwt.JwtConf               `yaml:"jwt" json:"jwt"`
	Mongo     mongo.MongoConf           `yaml:"mongo" json:"mongo"`
	Redis     redis.RedisConf           `yaml:"redis" json:"redis"`
	SafeReiew safereview.SafeReviewConf `yaml:"safereview" json:"safereview"`
	Snowflake snowflake.SnowflakeConf   `yaml:"snowflake" json:"snowflake"`
	Storage   storage.StorageConf       `yaml:"storage" json:"storage"`
}

var Config Configuration

func LoadConfig(filePath string) *Configuration {
	config.LoadConfig(filePath, &Config)
	return &Config
}
