package minio

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"protodesign.cn/kcserver/utils/storage/base"
	"strings"
)

type client struct {
	config *base.ClientConfig
	client *minio.Client
}

func NewClient(config *base.ClientConfig) (base.Client, error) {
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
	base.DefaultBucket
	config *base.BucketConfig
	client *client
}

func (that *client) NewBucket(config *base.BucketConfig) base.Bucket {
	instance := &bucket{
		config: config,
		client: that,
	}
	instance.That = instance
	return instance
}

func (that *bucket) PubObject(objectName string, reader io.Reader, objectSize int64, contentType string) (*base.UploadInfo, error) {
	var uploadInfo *base.UploadInfo
	if contentType == "" {
		contentType = "application/octet-stream"
		splitRes := strings.Split(objectName, ".")
		if len(splitRes) > 1 {
			switch splitRes[len(splitRes)-1] {
			case "json":
				contentType = "application/json"
			}
		}
	}
	_, err := that.client.client.PutObject(
		context.Background(),
		that.config.BucketName,
		objectName,
		reader,
		objectSize,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return nil, err
	}
	return uploadInfo, nil
}
