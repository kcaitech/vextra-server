package config

import (
	"protodesign.cn/kcserver/common/config"
)

type Configuration struct {
	Jwt struct {
		Secret string `yaml:"secret"`
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
