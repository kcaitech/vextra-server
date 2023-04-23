package config

import (
	"protodesign.cn/kcserver/common/config"
	"protodesign.cn/kcserver/utils/storage/base"
)

type Configuration struct {
	Storage struct {
		Provider base.Provider `yaml:"provider"`
	} `yaml:"storage"`
	Minio base.Config `yaml:"minio"`
}

var Config Configuration

func LoadConfig() *Configuration {
	config.LoadConfig("common/storage/config/config.yaml", &Config)
	return &Config
}
