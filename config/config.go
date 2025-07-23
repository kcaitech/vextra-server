package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
	mongo "kcaitech.com/kcserver/providers/mongo"
	redis "kcaitech.com/kcserver/providers/redis"
	safereview "kcaitech.com/kcserver/providers/safereview"
	storage "kcaitech.com/kcserver/providers/storage"
)

type DBConfig struct {
	// DSN string `yaml:"url" json:"url"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	Host     string `yaml:"host" json:"host"`
	Port     int64  `yaml:"port" json:"port"`
	Database string `yaml:"database" json:"database"`
	Charset  string `yaml:"charset" json:"charset"`
}

func (db *DBConfig) DSN() string {
	// root:kKEIjksvnOOIjdZ6rtzE@tcp(mysql:3306)/kcserver?charset=utf8&parseTime=True&loc=Local
	return db.User + ":" + db.Password + "@tcp(" + db.Host + ":" + fmt.Sprint(db.Port) + ")/" + db.Database + "?charset=" + db.Charset + "&parseTime=True&loc=Local"
}

type BaseConfiguration struct {
	DB         DBConfig `yaml:"db" json:"db"`
	StorageUrl struct {
		Document string `yaml:"document" json:"document"`
		Attatch  string `yaml:"attatch" json:"attatch"`
	} `yaml:"storage_public_url" json:"storage_public_url"`
}

// var Config *BaseConfiguration = &BaseConfiguration{}

// func Init(conf *BaseConfiguration) {
// 	Config = conf
// }

type AuthServerConfig struct {
	APIAddr      string `yaml:"api_addr" json:"api_addr"`
	LoginURL     string `yaml:"login_url" json:"login_url"`
	ClientID     string `yaml:"client_id" json:"client_id"`
	ClientSecret string `yaml:"client_secret" json:"client_secret"`
}

type MiddlewareConfig struct {
	Cors     bool `yaml:"cors,omitempty" json:"cors,omitempty"`
	DebugLog bool `yaml:"debug_log,omitempty" json:"debug_log,omitempty"`
}

type Configuration struct {
	BaseConfiguration `yaml:",inline" json:",inline"`
	VersionServer     struct {
		Url               string `yaml:"url" json:"url"`
		MinUpdateInterval int    `yaml:"min_update_interval" json:"min_update_interval"`
		MinCmdCount       int    `yaml:"min_cmd_count" json:"min_cmd_count"`
	} `yaml:"doc_update_server" json:"doc_update_server"`

	Mongo      mongo.MongoConf           `yaml:"mongo" json:"mongo"`
	Redis      redis.RedisConf           `yaml:"redis" json:"redis"`
	SafeReview safereview.SafeReviewConf `yaml:"safe_review" json:"safe_review"`
	Storage    storage.Config            `yaml:"storage" json:"storage"`

	Middleware MiddlewareConfig `yaml:"middleware" json:"middleware"`

	AuthServer AuthServerConfig `yaml:"auth_server" json:"auth_server"`
}

func loadYamlConfig(filePath string, config *Configuration) error {

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

func LoadYamlFile(filePath string) (*Configuration, error) {
	config := &Configuration{}
	err := loadYamlConfig(filePath, config)
	confirmConfig(config)
	return config, err
}
