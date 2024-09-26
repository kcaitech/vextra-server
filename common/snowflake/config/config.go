package config

import (
	"kcaitech.com/kcserver/common/config"
)

type Configuration struct {
	Snowflake struct {
		WorkerId int64 `yaml:"workerId"`
	} `yaml:"snowflake"`
}

var Config Configuration

func LoadConfig(filePath string) *Configuration {
	if filePath == "" {
		filePath = "common/snowflake/config/config.yaml"
	}
	config.LoadConfig(filePath, &Config)
	return &Config
}
