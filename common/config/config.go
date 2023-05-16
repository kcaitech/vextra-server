package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type BaseConfiguration struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	DB struct {
		DSN string `yaml:"dsn"`
	} `yaml:"db"`
}

func LoadConfig(filePath string, config any) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	err = yaml.Unmarshal(content, config)
	if err != nil {
		log.Fatalf("配置文件解析失败: %v", err)
	}
}
