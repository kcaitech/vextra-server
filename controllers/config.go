package controllers

import (
	"kcaitech.com/kcserver/common/config"
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
	DocumentVersionServer struct {
		Host string `yaml:"host"`
	} `yaml:"documentVersionServer"`
}

var Config Configuration

func LoadConfig(filePath string) *Configuration {
	config.LoadConfig(filePath, &Config)
	return &Config
}
