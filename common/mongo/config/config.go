package config

import (
	"kcaitech.com/kcserver/common/config"
)

type MongoConf struct {
	Uri string `yaml:"uri" json:"uri"`
	Db  string `yaml:"db" json:"db"`
}

type Configuration struct {
	Mongo MongoConf `yaml:"mongo" json:"mongo"`
}

var Config Configuration

func LoadConfig(filePath string) *Configuration {
	if filePath == "" {
		filePath = "common/mongo/config/config.yaml"
	}
	config.LoadConfig(filePath, &Config)
	return &Config
}
