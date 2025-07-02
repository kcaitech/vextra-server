package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	AuthDB struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Database string `json:"database"`
	} `json:"authdb"`
}

func LoadYamlFile(filePath string) (*Config, error) {
	var conf Config
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("读取配置文件失败: %v", err)
		return nil, err
	}
	err = yaml.Unmarshal(content, &conf)
	if err != nil {
		log.Printf("配置文件解析失败: %v", err)
		return nil, err
	}
	return &conf, nil
}
