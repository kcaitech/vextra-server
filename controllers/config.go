package controllers

import (
	"errors"
	"os"

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
	VersionServer struct {
		Url               string `yaml:"url" json:"url"`
		MinUpdateInterval int    `yaml:"min_update_interval" json:"min_update_interval"`
	} `yaml:"version_server" json:"version_server"`
	Svg2Png struct {
		Url string `yaml:"url" json:"url"`
	} `yaml:"svg2png" json:"svg2png"`

	Jwt       jwt.JwtConf               `yaml:"jwt" json:"jwt"`
	Mongo     mongo.MongoConf           `yaml:"mongo" json:"mongo"`
	Redis     redis.RedisConf           `yaml:"redis" json:"redis"`
	SafeReiew safereview.SafeReviewConf `yaml:"safereview" json:"safe_review"`
	Snowflake snowflake.SnowflakeConf   `yaml:"snowflake" json:"snowflake"`
	Storage   storage.StorageConf       `yaml:"storage" json:"storage"`
}

var Config Configuration

func LoadConfig(filePath string) (*Configuration, error) {
	err := config.LoadJsonConfig(filePath, &Config)
	return &Config, err
}

func LoadConfigEnv(env string) (*Configuration, error) {
	content := os.Getenv(env)
	if content == "" {
		return &Config, errors.New("no " + env)
	}
	config.LoadJsonContent(content, &Config)
	return &Config, nil
}
