package config

import (
	"protodesign.cn/kcserver/common/config"
	"protodesign.cn/kcserver/utils/storage/minio"
)

type Configuration struct {
	config.BaseConfiguration `yaml:",inline"`
	minio.ClientConfig       `yaml:",inline"`
	minio.BucketConfig       `yaml:",inline"`
}

var Config Configuration

func LoadConfig() *Configuration {
	config.LoadConfig("config/config.yaml", &Config)
	return &Config
}
