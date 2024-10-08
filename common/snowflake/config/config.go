package config

import (
	"kcaitech.com/kcserver/common/config"
)

type SnowflakeConf struct {
	WorkerId int64 `yaml:"workerId" json:"workerId"`
}

type Configuration struct {
	Snowflake SnowflakeConf `yaml:"snowflake" json:"snowflake"`
}

var Config Configuration

func LoadConfig(filePath string) *Configuration {
	if filePath == "" {
		filePath = "common/snowflake/config/config.yaml"
	}
	config.LoadConfig(filePath, &Config)
	return &Config
}
