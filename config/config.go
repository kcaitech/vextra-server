package config

import (
	"encoding/json"
	"errors"
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

// var Config *BaseConfiguration = &BaseConfiguration{}

// func Init(conf *BaseConfiguration) {
// 	Config = conf
// }

func loadYamlConfig(filePath string, config any) error {

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
func loadJsonConfig(filePath string, config any) error {

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
func loadJsonContent(content string, config any) error {

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

type Configuration struct {
	BaseConfiguration `yaml:",inline" json:",inline"`
	// Wx                struct {
	// 	Appid  string `yaml:"appid" json:"appid"`
	// 	Secret string `yaml:"secret" json:"secret"`
	// } `yaml:"wx" json:"wx"`
	// WxMp struct {
	// 	Appid  string `yaml:"appid" json:"appid"`
	// 	Secret string `yaml:"secret" json:"secret"`
	// } `yaml:"wxMp" json:"wxMp"`
	VersionServer struct {
		Url               string `yaml:"url" json:"url"`
		MinUpdateInterval int    `yaml:"min_update_interval" json:"min_update_interval"`
		MinCmdCount       int    `yaml:"min_cmd_count" json:"min_cmd_count"`
	} `yaml:"version_server" json:"version_server"`
	Svg2Png struct {
		Url string `yaml:"url" json:"url"`
	} `yaml:"svg2png" json:"svg2png"`

	// Jwt       jwt.JwtConf               `yaml:"jwt" json:"jwt"`
	Mongo     mongo.MongoConf           `yaml:"mongo" json:"mongo"`
	Redis     redis.RedisConf           `yaml:"redis" json:"redis"`
	SafeReiew safereview.SafeReviewConf `yaml:"safe_review" json:"safe_review"`
	// Snowflake snowflake.SnowflakeConf   `yaml:"snowflake" json:"snowflake"`
	Storage storage.StorageConf `yaml:"storage" json:"storage"`

	DefaultRoute bool `yaml:"default_route,omitempty" json:"default_route,omitempty"`
	DetailedLog  bool `yaml:"detailed_log,omitempty" json:"detailed_log,omitempty"`
	AllowCors    bool `yaml:"allow_cors,omitempty" json:"allow_cors,omitempty"`

	AuthServerURL    string `yaml:"auth_server_url,omitempty" json:"auth_server_url,omitempty"`
	AuthClientID     string `yaml:"auth_client_id,omitempty" json:"auth_client_id,omitempty"`
	AuthClientSecret string `yaml:"auth_client_secret,omitempty" json:"auth_client_secret,omitempty"`
	AuthCallbackURL  string `yaml:"auth_callback_url,omitempty" json:"auth_callback_url,omitempty"`
}

func LoadYamlFile(filePath string) (*Configuration, error) {
	var Config Configuration
	err := loadYamlConfig(filePath, &Config)
	return &Config, err
}

func LoadJsonFile(filePath string) (*Configuration, error) {
	var Config Configuration
	err := loadJsonConfig(filePath, &Config)
	return &Config, err
}

func LoadJsonEnv(env string) (*Configuration, error) {
	var Config Configuration
	content := os.Getenv(env)
	if content == "" {
		return &Config, errors.New("no " + env)
	}
	loadJsonContent(content, &Config)
	return &Config, nil
}
