package config

import (
	"protodesign.cn/kcserver/common/config"
)

type Configuration struct {
	Snowflake struct {
		WorkerId int64 `yaml:"workerId"`
	} `yaml:"snowflake"`
}

var Config Configuration

func LoadConfig() *Configuration {
	config.LoadConfig("common/snowflake/config/config.yaml", &Config)
	return &Config
}
