package config

import (
	"protodesign.cn/kcserver/common/config"
)

type Configuration struct {
	config.BaseConfiguration `yaml:",inline"`
	Wx                       struct {
		Appid  string `yaml:"appid"`
		Secret string `yaml:"secret"`
	} `yaml:"wx"`
	WxMp struct {
		Appid  string `yaml:"appid"`
		Secret string `yaml:"secret"`
	} `yaml:"wxMp"`
}

var Config Configuration

func LoadConfig() *Configuration {
	config.LoadConfig("config/config.yaml", &Config)
	return &Config
}
