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
	Provider           Provider `yaml:"provider"`
	Endpoint           string   `yaml:"endpoint"`
	Region             string   `yaml:"region"`
	AccessKeyID        string   `yaml:"accessKeyID"`
	SecretAccessKey    string   `yaml:"secretAccessKey"`
	StsAccessKeyID     string   `yaml:"stsAccessKeyID"`
	StsSecretAccessKey string   `yaml:"stsSecretAccessKey"`
}

type Bucket interface {
	PubObject(objectName string, reader io.Reader, objectSize int64, contentType string) (*UploadInfo, error)
	PubObjectByte(objectName string, content []byte) (*UploadInfo, error)
	GenerateAccessKey(authPath string, authOp int, expires int) (*AccessKeyValue, error)
}

type BucketConfig struct {
	BucketName      string `yaml:"bucketName"`
	FilesBucketName string `yaml:"filesBucketName"`
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

type AccessKeyValue struct {
	AccessKey       string `json:"access_key"`
	SecretAccessKey string `json:"secret_access_key"`
	SessionToken    string `json:"session_token"`
	SignerType      int    `json:"signer_type"`
}

const (
	AuthOpGetObject  = 1 << 0
	AuthOpPutObject  = 1 << 1
	AuthOpDelObject  = 1 << 2
	AuthOpListObject = 1 << 3
	AuthOpAll        = AuthOpGetObject | AuthOpPutObject | AuthOpDelObject | AuthOpListObject
)

func (that *DefaultBucket) GenerateAccessKey(authPath string, authOp int, expires int) (*AccessKeyValue, error) {
	return nil, errors.New("generateAccessKey方法未实现")
}
