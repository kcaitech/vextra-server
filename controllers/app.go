package controllers

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"kcaitech.com/kcserver/common/gin/auth"
	"kcaitech.com/kcserver/common/gin/response"
	"kcaitech.com/kcserver/common/services"
)

// GetAppVersionList 获取APP版本列表
func GetAppVersionList(c *gin.Context) {
	appVersionService := services.NewAppVersionService()
	result := appVersionService.FindAll()
	response.Success(c, result)
}

// GetLatestAppVersion 获取最新的版本信息
func GetLatestAppVersion(c *gin.Context) {
	userId, _ := auth.GetUserId(c)

	appVersionService := services.NewAppVersionService()
	result := appVersionService.GetLatest(userId)
	response.Success(c, result)
}

type Package struct {
	Version string `yaml:"version"`
}

func LoadPackageVersion() string {
	content, err := os.ReadFile("package.yaml")
	if err != nil {
		log.Fatalf("load package.yaml fail %v", err)
	}
	config := &Package{}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		log.Fatalf("unmarshal package.yaml fail %v", err)
	}
	return config.Version
}

var version = LoadPackageVersion()

func GetAppVersion(c *gin.Context) {
	response.Success(c, version)
}
