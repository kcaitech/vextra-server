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
	Jwt struct {
		Ref string `yaml:"ref"`
	} `yaml:"jwt"`
	MongoDb struct {
		Ref string `yaml:"ref"`
	} `yaml:"mongodb"`
	Redis struct {
		Ref string `yaml:"ref"`
	} `yaml:"redis"`
	SafeReiew struct {
		Ref string `yaml:"ref"`
	} `yaml:"safereview"`
	Snowflake struct {
		Ref string `yaml:"ref"`
	} `yaml:"snowflake"`
	Storage struct {
		Ref string `yaml:"ref"`
	} `yaml:"storage"`
}

var Config Configuration

func LoadConfig(filePath string) *Configuration {
	config.LoadConfig(filePath, &Config)
	return &Config
}
