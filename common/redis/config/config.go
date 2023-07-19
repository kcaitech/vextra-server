package config

import (
	"protodesign.cn/kcserver/common/config"
)

type Configuration struct {
	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		Db       int    `yaml:"db"`
	} `yaml:"redis"`
}

var Config Configuration

func LoadConfig(filePath string) *Configuration {
	if filePath == "" {
		filePath = "common/redis/config/config.yaml"
	}
	config.LoadConfig(filePath, &Config)
	return &Config
}
