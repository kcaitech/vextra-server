package config

import (
	"kcaitech.com/kcserver/common/config"
)

type RedisConf struct {
	Addr          string   `yaml:"addr" json:"addr"`
	Password      string   `yaml:"password" json:"password"`
	Db            int      `yaml:"db" json:"db"`
	Sentinel      bool     `yaml:"sentinel" json:"sentinel"`
	SentinelAddrs []string `yaml:"sentinelAddrs" json:"sentinelAddrs"`
	MasterName    string   `yaml:"masterName" json:"masterName"`
}

type Configuration struct {
	Redis RedisConf `yaml:"redis" json:"redis"`
}

func LoadConfig(filePath string) *Configuration {
	var Config Configuration
	if filePath == "" {
		filePath = "common/redis/config/config.yaml"
	}
	config.LoadConfig(filePath, &Config)
	return &Config
}
