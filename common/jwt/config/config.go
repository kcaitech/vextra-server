package config

import (
	"kcaitech.com/kcserver/common/config"
)

type Configuration struct {
	Jwt struct {
		Secret     string `yaml:"secret"`
		ExpireHour int64  `yaml:"expire_hour"`
	} `yaml:"jwt"`
}

var Config Configuration

func LoadConfig(filePath string) *Configuration {
	if filePath == "" {
		filePath = "common/jwt/config/config.yaml"
	}
	config.LoadConfig(filePath, &Config)
	return &Config
}
