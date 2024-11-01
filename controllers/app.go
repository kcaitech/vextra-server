package controllers

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"kcaitech.com/kcserver/common/gin/response"
)

// GetAppVersionList 获取APP版本列表
// func GetAppVersionList(c *gin.Context) {
// 	appVersionService := services.NewAppVersionService()
// 	result := appVersionService.FindAll()
// 	response.Success(c, result)
// }

// GetLatestAppVersion 获取最新的版本信息
// func GetLatestAppVersion(c *gin.Context) {
// 	userId, _ := auth.GetUserId(c)

// 	appVersionService := services.NewAppVersionService()
// 	result := appVersionService.GetLatest(userId)
// 	response.Success(c, result)
// }

type Package struct {
	Version string `yaml:"version"`
}

func LoadPackageVersion() *string {
	var def = ""
	content, err := os.ReadFile("package.yaml")
	if err != nil {
		log.Printf("load package.yaml fail %v", err)
		return &def
	}
	config := &Package{}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		log.Printf("unmarshal package.yaml fail %v", err)
		return &def
	}
	return &config.Version
}

var version *string

func GetAppVersion(c *gin.Context) {
	if version == nil {
		version = LoadPackageVersion()
	}
	response.Success(c, version)
}

var index_hash *[]byte

func GetIndexHash(c *gin.Context) {
	if index_hash == nil {
		content, err := os.ReadFile("/app/html/index.hash")
		if err != nil {
			log.Printf("load index.hash fail %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		index_hash = &content
	}
	c.String(http.StatusOK, string(*index_hash))
}
