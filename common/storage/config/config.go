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
	S3    base.Config `yaml:"s3"`
	Oss   base.Config `yaml:"oss"`
}

var Config Configuration

func LoadConfig(filePath string) *Configuration {
	if filePath == "" {
		filePath = "common/storage/config/config.yaml"
	}
	config.LoadConfig(filePath, &Config)
	return &Config
}
