package config

import (
	"protodesign.cn/kcserver/common/config"
	"protodesign.cn/kcserver/common/safereview/base"
)

type Configuration struct {
	Provider base.Provider `yaml:"provider"`
	Ali      struct {
		AccessKeyId     string `yaml:"accessKeyId"`
		AccessKeySecret string `yaml:"accessKeySecret"`
		RegionId        string `yaml:"regionId"`
		Endpoint        string `yaml:"endpoint"`
	} `yaml:"ali"`
}

var Config Configuration

func LoadConfig(filePath string) *Configuration {
	if filePath == "" {
		filePath = "common/safereview/config/config.yaml"
	}
	config.LoadConfig(filePath, &Config)
	return &Config
}
