package main

// Config 迁移配置结构
type Config struct {
	Source struct {
		MySQL struct {
			User     string `json:"user"`
			Password string `json:"password"`
			Host     string `json:"host"`
			Port     int    `json:"port"`
			Database string `json:"database"`
		} `json:"mysql"`
		Mongo struct {
			URL string `json:"url"`
			DB  string `json:"db"`
		} `json:"mongo"`
		GenerateApiUrl string `json:"generateApiUrl"`
		Storage        struct {
			Provider string `json:"provider"`
			Minio    struct {
				Endpoint     string `json:"endpoint"`
				Region       string `json:"region"`
				AccessKey    string `json:"accessKeyID"`
				SecretKey    string `json:"secretAccessKey"`
				StsAccessKey string `json:"stsAccessKeyID"`
				StsSecretKey string `json:"stsSecretAccessKey"`
				Bucket       string `json:"bucketName"`
				FilesBucket  string `json:"filesBucketName"`
			} `json:"minio"`
			S3 struct {
				Endpoint    string `json:"endpoint"`
				Region      string `json:"region"`
				AccessKey   string `json:"accessKeyID"`
				SecretKey   string `json:"secretAccessKey"`
				AccountId   string `json:"accountID"`
				RoleName    string `json:"roleName"`
				Bucket      string `json:"bucketName"`
				FilesBucket string `json:"filesBucketName"`
			} `json:"s3"`
			OSS struct {
				Endpoint    string `json:"endpoint"`
				StsEndpoint string `json:"stsEndpoint"`
				Region      string `json:"region"`
				AccessKey   string `json:"accessKeyID"`
				SecretKey   string `json:"secretAccessKey"`
				AccountId   string `json:"accountId"`
				RoleName    string `json:"roleName"`
				Bucket      string `json:"bucketName"`
				FilesBucket string `json:"filesBucketName"`
			} `json:"oss"`
		} `json:"storage"`
	} `json:"source"`
	Target struct {
		MySQL struct {
			User     string `json:"user"`
			Password string `json:"password"`
			Host     string `json:"host"`
			Port     int    `json:"port"`
			Database string `json:"database"`
		} `json:"mysql"`
		Mongo struct {
			URL string `json:"url"`
			DB  string `json:"db"`
		} `json:"mongo"`
	} `json:"target"`
	Auth struct {
		MySQL struct {
			User     string `json:"user"`
			Password string `json:"password"`
			Host     string `json:"host"`
			Port     int    `json:"port"`
			Database string `json:"database"`
		} `json:"mysql"`
	} `json:"auth"`
}

// NewWeixinUser 微信用户结构
type NewWeixinUser struct {
	UserID  string `json:"user_id" gorm:"primarykey"`
	UnionID string `json:"union_id" gorm:"unique"`
}
