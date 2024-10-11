package config

import (
	"encoding/json"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type BaseConfiguration struct {
	Server struct {
		Port string `yaml:"port" json:"port"`
	} `yaml:"server" json:"server"`
	DB struct {
		DSN string `yaml:"dsn" json:"dsn"`
	} `yaml:"db" json:"db"`
	StorageHost struct {
		Document string `yaml:"document" json:"document"`
		Attatch  string `yaml:"attatch" json:"attatch"`
	} `yaml:"storage_host" json:"storage_host"`
}

var Config *BaseConfiguration = &BaseConfiguration{}

func Init(conf *BaseConfiguration) {
	Config = conf
}

func LoadConfig(filePath string, config any) error {

	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
		return err
	}

	err = yaml.Unmarshal(content, config)
	if err != nil {
		log.Fatalf("配置文件解析失败: %v", err)
		return err
	}
	return nil
}
func LoadJsonConfig(filePath string, config any) error {

	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
		return err
	}

	err = json.Unmarshal(content, config)
	if err != nil {
		log.Fatalf("配置文件解析失败: %v", err)
		return err
	}
	return nil
}
func LoadJsonContent(content string, config any) error {

	// content, err := os.ReadFile(filePath)
	// if err != nil {
	// 	log.Fatalf("读取配置文件失败: %v", err)
	// }

	err := json.Unmarshal([]byte(content), config)
	if err != nil {
		log.Fatalf("配置文件解析失败: %v", err)
		return err
	}
	return nil
}
