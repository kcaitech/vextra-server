package config

import (
	"kcaitech.com/kcserver/common/config"
	"kcaitech.com/kcserver/utils/storage/base"
)

type StorageConf struct {
	Provider base.Provider `yaml:"provider" json:"provider"`
	Minio    base.Config   `yaml:"minio" json:"minio"`
	S3       base.Config   `yaml:"s3" json:"s3"`
	Oss      base.Config   `yaml:"oss" json:"oss"`
	// Hosts    struct {
	// 	Main  string `yaml:"main" json:"main"`
	// 	Files string `yaml:"files" json:"files"`
	// } `yaml:"hosts" json:"hosts"`
}

type Configuration struct {
	Storage StorageConf `yaml:"storage" json:"storage"`
}

var Config Configuration

func LoadConfig(filePath string) *Configuration {
	if filePath == "" {
		filePath = "common/storage/config/config.yaml"
	}
	config.LoadConfig(filePath, &Config)
	return &Config
}
