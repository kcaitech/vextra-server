package minio

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type client struct {
	config *ClientConfig
	client *minio.Client
}

type ClientConfig struct {
	Endpoint        string `yaml:"endpoint"`
	Port            string `yaml:"port"`
	AccessKeyID     string `yaml:"accessKeyID"`
	SecretAccessKey string `yaml:"secretAccessKey"`
}

func NewClient(config *ClientConfig) (*client, error) {
	c, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}
	return &client{
		config: config,
		client: c,
	}, nil
}

type bucket struct {
	config *BucketConfig
}

type BucketConfig struct {
	BucketName string `yaml:"bucketName"`
	Region     string `yaml:"region"`
}

func (that *client) NewBucket(config *BucketConfig) *bucket {
	return &bucket{
		config: config,
	}
}
