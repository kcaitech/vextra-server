package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type DBConfig struct {
	// DSN string `yaml:"url" json:"url"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	Host     string `yaml:"host" json:"host"`
	Port     int64  `yaml:"port" json:"port"`
	Database string `yaml:"database" json:"database"`
}

func (db *DBConfig) DSN() string {
	// root:kKEIjksvnOOIjdZ6rtzE@tcp(mysql:3306)/kcserver?charset=utf8&parseTime=True&loc=Local
	return db.User + ":" + db.Password + "@tcp(" + db.Host + ":" + fmt.Sprint(db.Port) + ")/" + db.Database + "?charset=utf8&parseTime=True&loc=Local"
}

type BaseConfiguration struct {
	DB         DBConfig `yaml:"db" json:"db"`
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
		log.Printf("读取配置文件失败: %v", err)
		return err
	}

	err = yaml.Unmarshal(content, config)
	if err != nil {
		log.Printf("配置文件解析失败: %v", err)
		return err
	}
	return nil
}
func LoadJsonConfig(filePath string, config any) error {

	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("读取配置文件失败: %v", err)
		return err
	}

	err = json.Unmarshal(content, config)
	if err != nil {
		log.Printf("配置文件解析失败: %v", err)
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
		log.Printf("配置文件解析失败: %v", err)
		return err
	}
	return nil
}
