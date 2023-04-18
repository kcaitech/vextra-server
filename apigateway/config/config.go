package config

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
)

type Configuration struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	DB struct {
		DSN string `yaml:"dsn"`
	} `yaml:"db"`
}

var Config Configuration

func LoadConfig() {
	file, err := ioutil.ReadFile("config/config.yaml")
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	err = yaml.Unmarshal(file, &Config)
	if err != nil {
		log.Fatalf("配置文件解析失败: %v", err)
	}
}
