package config

import (
	"protodesign.cn/kcserver/common/config"
)

type Configuration struct {
	Mongo struct {
		Uri string `yaml:"uri"`
		Db  string `yaml:"db"`
	} `yaml:"mongo"`
}

var Config Configuration

func LoadConfig(filePath string) *Configuration {
	if filePath == "" {
		filePath = "common/mongo/config/config.yaml"
	}
	config.LoadConfig(filePath, &Config)
	return &Config
}
