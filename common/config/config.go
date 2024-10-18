package config

import (
	"encoding/json"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type BaseConfiguration struct {
	DB struct {
		DSN string `yaml:"url" json:"url"`
	} `yaml:"db" json:"db"`
	StorageUrl struct {
		Document string `yaml:"document" json:"document"`
		Attatch  string `yaml:"attatch" json:"attatch"`
	} `yaml:"storage_url" json:"storage_url"`
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
