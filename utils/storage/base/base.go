package base

import (
	"bytes"
	"errors"
	"io"
)

type Client interface {
	NewBucket(config *BucketConfig) Bucket
}

type ClientConfig struct {
	Provider        Provider `yaml:"provider"`
	Endpoint        string   `yaml:"endpoint"`
	AccessKeyID     string   `yaml:"accessKeyID"`
	SecretAccessKey string   `yaml:"secretAccessKey"`
}

type Bucket interface {
	PubObject(objectName string, reader io.Reader, objectSize int64, contentType string) (*UploadInfo, error)
	PubObjectByte(objectName string, content []byte) (*UploadInfo, error)
}

type BucketConfig struct {
	BucketName string `yaml:"bucketName"`
	Region     string `yaml:"region"`
}

type Config struct {
	ClientConfig `yaml:",inline"`
	BucketConfig `yaml:",inline"`
}

type UploadInfo struct {
}

type DefaultBucket struct {
	That Bucket
}

func (that *DefaultBucket) PubObject(objectName string, reader io.Reader, objectSize int64, contentType string) (*UploadInfo, error) {
	return nil, errors.New("PubObject方法未实现")
}

func (that *DefaultBucket) PubObjectByte(objectName string, content []byte) (*UploadInfo, error) {
	return that.That.PubObject(objectName, bytes.NewReader(content), int64(len(content)), "")
}
