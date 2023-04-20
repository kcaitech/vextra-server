package config

import (
	"protodesign.cn/kcserver/common/config"
)

type Configuration struct {
	config.BaseConfiguration `yaml:",inline"`
}

var Config Configuration

func LoadConfig() *Configuration {
	config.LoadConfig("config/config.yaml", &Config)
	return &Config
}
