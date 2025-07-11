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
	Client        Client
	Bucket        Bucket
	AttatchBucket Bucket
}

func NewStoraageClient(config *Config) (*StorageClient, error) {
	var client Client
	var err error
	var bucketConfig = config.BucketConfig
	switch config.Provider {
	case MINIO:
		client, err = NewMinioClient(&config.ClientConfig)
	case S3:
		client, err = NewS3Client(&config.ClientConfig)
	case OSS:
		client, err = NewOSSClient(&config.ClientConfig)
	default:
		return nil, errors.New("不支持的provider")
	}

	if err != nil {
		return nil, err
	}

	return &StorageClient{
		Client: client,
		Bucket: client.NewBucket(&BucketConfig{
			DocumentBucket: bucketConfig.DocumentBucket,
		}),
		AttatchBucket: client.NewBucket(&BucketConfig{
			DocumentBucket: bucketConfig.AttatchBucket,
		}),
	}, nil
}
