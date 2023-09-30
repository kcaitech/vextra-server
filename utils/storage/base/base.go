package base

import (
	"bytes"
	"errors"
	"io"
	"protodesign.cn/kcserver/utils/my_map"
	"sync"
)

type Client interface {
	NewBucket(config *BucketConfig) Bucket
}

type ClientConfig struct {
	Provider        Provider `yaml:"provider"`
	Endpoint        string   `yaml:"endpoint"`
	Region          string   `yaml:"region"`
	AccessKeyID     string   `yaml:"accessKeyID"`
	SecretAccessKey string   `yaml:"secretAccessKey"`

	// minio sts
	StsAccessKeyID     string `yaml:"stsAccessKeyID"`
	StsSecretAccessKey string `yaml:"stsSecretAccessKey"`

	// s3 sts
	StsEndpoint string `yaml:"stsEndpoint"`
	AccountId   string `yaml:"accountId"`
	RoleName    string `yaml:"roleName"`
}

type PutObjectInput struct {
	ObjectName  string
	Reader      io.Reader
	ObjectSize  int64
	ContentType string
}

type Bucket interface {
	PutObject(putObjectInput *PutObjectInput) (*UploadInfo, error)
	PutObjectByte(objectName string, content []byte) (*UploadInfo, error)
	PutObjectList(putObjectInputList []*PutObjectInput) ([]*UploadInfo, []error)
	GenerateAccessKey(authPath string, authOp int, expires int, roleSessionName string) (*AccessKeyValue, error)
	CopyObject(srcPath string, destPath string) (*UploadInfo, error)
	CopyDirectory(srcDirPath string, destDirPath string) (*UploadInfo, error)
	GetObjectInfo(objectName string) (*ObjectInfo, error)
	GetObject(objectName string) ([]byte, error)
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
	VersionID string `json:"versionId"`
}

type ObjectInfo struct {
	VersionID string `json:"versionId"`
}

type DefaultBucket struct {
	That Bucket
}

func (that *DefaultBucket) PubObject(putObjectInput *PutObjectInput) (*UploadInfo, error) {
	return nil, errors.New("PubObject方法未实现")
}

func (that *DefaultBucket) PutObjectByte(objectName string, content []byte) (*UploadInfo, error) {
	return that.That.PutObject(&PutObjectInput{
		ObjectName:  objectName,
		Reader:      bytes.NewReader(content),
		ObjectSize:  int64(len(content)),
		ContentType: "",
	})
}

func (that *DefaultBucket) PutObjectList(putObjectInputList []*PutObjectInput) ([]*UploadInfo, []error) {
	uploadInfoMap := my_map.NewSyncMap[int, *UploadInfo]()
	errorMap := my_map.NewSyncMap[int, error]()
	uploadWaitGroup := sync.WaitGroup{}
	for index, putObjectInput := range putObjectInputList {
		uploadWaitGroup.Add(1)
		go func(index int, putObjectInput *PutObjectInput) {
			defer uploadWaitGroup.Done()
			result, err := that.That.PutObject(putObjectInput)
			uploadInfoMap.Set(index, result)
			errorMap.Set(index, err)
		}(index, putObjectInput)
	}
	uploadWaitGroup.Wait()
	uploadInfoList := make([]*UploadInfo, 0, len(putObjectInputList))
	errorList := make([]error, 0, len(putObjectInputList))
	for index := 0; index < len(putObjectInputList); index++ {
		uploadInfo, _ := uploadInfoMap.Get(index)
		err, _ := errorMap.Get(index)
		uploadInfoList = append(uploadInfoList, uploadInfo)
		errorList = append(errorList, err)
	}
	return uploadInfoList, errorList
}

func (that *DefaultBucket) CopyObject(srcPath string, destPath string) (*UploadInfo, error) {
	return nil, errors.New("CopyObject方法未实现")
}

func (that *DefaultBucket) CopyDirectory(srcDirPath string, destDirPath string) (*UploadInfo, error) {
	return nil, errors.New("CopyDirectory方法未实现")
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

func (that *DefaultBucket) GenerateAccessKey(authPath string, authOp int, expires int, roleSessionName string) (*AccessKeyValue, error) {
	return nil, errors.New("generateAccessKey方法未实现")
}
