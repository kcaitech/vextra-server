package config

import "kcaitech.com/kcserver/common/config"

type JwtConf struct {
	Secret     string `yaml:"secret" json:"secret"`
	ExpireHour int64  `yaml:"expire_hour" json:"expire_hour"`
}

type Configuration struct {
	Jwt JwtConf `yaml:"jwt" json:"jwt"`
}

// var Config Configuration

func LoadConfig(filePath string) *Configuration {
	var Config Configuration
	if filePath == "" {
		filePath = "common/jwt/config/config.yaml"
	}
	config.LoadConfig(filePath, &Config)
	return &Config
}
