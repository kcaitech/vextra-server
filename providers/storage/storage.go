package storage

import (
	"errors"
)

type StorageConf struct {
	Provider Provider `yaml:"provider" json:"provider"`
	Minio    Config   `yaml:"minio" json:"minio"`
	S3       Config   `yaml:"s3" json:"s3"`
	Oss      Config   `yaml:"oss" json:"oss"`
}

type StorageClient struct {
	Client      Client
	Bucket      Bucket
	FilesBucket Bucket
}

func NewStoraageClient(config *StorageConf) (*StorageClient, error) {
	var client Client
	var err error
	var bucketConfig BucketConfig
	switch config.Provider {
	case MINIO:
		client, err = NewMinioClient(&config.Minio.ClientConfig)
		bucketConfig = config.Minio.BucketConfig
	case S3:
		client, err = NewS3Client(&config.S3.ClientConfig)
		bucketConfig = config.S3.BucketConfig
	case OSS:
		client, err = NewOSSClient(&config.Oss.ClientConfig)
		bucketConfig = config.Oss.BucketConfig
	default:
		return nil, errors.New("不支持的provider")
	}

	if err != nil {
		return nil, err
	}

	return &StorageClient{
		Client: client,
		Bucket: client.NewBucket(&BucketConfig{
			BucketName: bucketConfig.BucketName,
		}),
		FilesBucket: client.NewBucket(&BucketConfig{
			BucketName: bucketConfig.FilesBucketName,
		}),
	}, nil
}

// var Client base.Client
// var Bucket base.Bucket
// var FilesBucket base.Bucket

// func Init(conf *config.StorageConf) error {
// 	// conf := config.LoadConfig(filePath)

// 	var providerConf base.Config
// 	switch conf.Provider {
// 	case base.MINIO:
// 		providerConf = conf.Minio
// 	case base.S3:
// 		providerConf = conf.S3
// 	case base.OSS:
// 		providerConf = conf.Oss
// 	default:
// 		return errors.New("不支持的provider")
// 	}

// 	providerConf.ClientConfig.Provider = conf.Provider

// 	var err error
// 	if Client, err = storage.NewClient(&providerConf.ClientConfig); err != nil {
// 		return err
// 	}
// 	Bucket = Client.NewBucket(&base.BucketConfig{
// 		BucketName: providerConf.BucketName,
// 	})
// 	FilesBucket = Client.NewBucket(&base.BucketConfig{
// 		BucketName: providerConf.FilesBucketName,
// 	})
// 	return nil
// }
