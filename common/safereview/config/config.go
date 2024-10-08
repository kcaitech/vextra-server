package config

import (
	"kcaitech.com/kcserver/common/config"
	"kcaitech.com/kcserver/common/safereview/base"
)

type SafeReviewConf struct {
	Provider base.Provider `yaml:"provider" json:"provider"`
	Ali      struct {
		AccessKeyId     string `yaml:"accessKeyId" json:"accessKeyId"`
		AccessKeySecret string `yaml:"accessKeySecret" json:"accessKeySecret"`
		RegionId        string `yaml:"regionId" json:"regionId"`
		Endpoint        string `yaml:"endpoint" json:"endpoint"`
	} `yaml:"ali" json:"ali"`
	Baidu struct {
		AppId     string `yaml:"appId" json:"appId"`
		ApiKey    string `yaml:"apiKey" json:"apiKey"`
		SecretKey string `yaml:"secretKey" json:"secretKey"`
	} `yaml:"baidu" json:"baidu"`
}

type Configuration struct {
	SafeReview SafeReviewConf `yaml:"safereview" json:"safereview"`
}

var Config Configuration

func LoadConfig(filePath string) *Configuration {
	if filePath == "" {
		filePath = "common/safereview/config/config.yaml"
	}
	config.LoadConfig(filePath, &Config)
	return &Config
}
