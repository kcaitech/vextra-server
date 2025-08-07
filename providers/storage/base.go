/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package storage

import (
	"bytes"
	"errors"
	"io"
)

type Provider string

const (
	MINIO Provider = "minio"
	S3    Provider = "s3"
	OSS   Provider = "oss"
)

type Client interface {
	NewBucket(config *BucketConfig) Bucket
}

type ClientConfig struct {
	// Provider        Provider `yaml:"provider" json:"provider"`
	Endpoint        string `yaml:"endpoint" json:"endpoint"`
	Region          string `yaml:"region" json:"region"`
	AccessKeyID     string `yaml:"accessKeyID" json:"accessKeyID"`
	SecretAccessKey string `yaml:"secretAccessKey" json:"secretAccessKey"`

	// minio sts
	StsAccessKeyID     string `yaml:"stsAccessKeyID" json:"stsAccessKeyID"`
	StsSecretAccessKey string `yaml:"stsSecretAccessKey" json:"stsSecretAccessKey"`

	// s3/oss sts
	StsEndpoint string `yaml:"stsEndpoint" json:"stsEndpoint"`
	AccountId   string `yaml:"accountId" json:"accountId"`
	RoleName    string `yaml:"roleName" json:"roleName"`
}

type PutObjectInput struct {
	ObjectName  string
	Reader      io.Reader
	ObjectSize  int64
	ContentType string
}

type Bucket interface {
	GetConfig() *Config
	PutObject(putObjectInput *PutObjectInput) (*UploadInfo, error)
	PutObjectByte(objectName string, content []byte, contentType string) (*UploadInfo, error)
	// PutObjectList(putObjectInputList []*PutObjectInput) ([]*UploadInfo, []error)
	GenerateAccessKey(authPath string, authOp int, expires int, roleSessionName string) (*AccessKeyValue, error)
	CopyObject(srcPath string, destPath string) (*UploadInfo, error)
	CopyDirectory(srcDirPath string, destDirPath string) (*UploadInfo, error)
	GetObjectInfo(objectName string) (*ObjectInfo, error)
	GetObject(objectName string) ([]byte, error)
	DeleteObject(objectName string) error
	ListObjects(prefix string) <-chan ObjectInfo
	// PresignedGetObject(objectName string, expires time.Duration, reqParams url.Values) (string, error)
}

type BucketConfig struct {
	DocumentBucket string `yaml:"documentBucket" json:"documentBucket"`
	AttatchBucket  string `yaml:"attatchBucket" json:"attatchBucket"`
}

type Config struct {
	Provider     Provider `yaml:"provider" json:"provider"`
	ClientConfig `yaml:",inline" json:",inline"`
	BucketConfig `yaml:",inline" json:",inline"`
}

type UploadInfo struct {
	VersionID string `json:"versionId"`
}

type ObjectInfo struct {
	Key       string
	Err       error
	Size      int64
	VersionID string
}

type DefaultBucket struct {
	That Bucket
}

func (that *DefaultBucket) GetConfig() (*Config, error) {
	return nil, errors.New("GetConfig方法未实现")
}

func (that *DefaultBucket) PubObject(putObjectInput *PutObjectInput) (*UploadInfo, error) {
	return nil, errors.New("PubObject方法未实现")
}

func (that *DefaultBucket) PutObjectByte(objectName string, content []byte, contentType string) (*UploadInfo, error) {
	return that.That.PutObject(&PutObjectInput{
		ObjectName:  objectName,
		Reader:      bytes.NewReader(content),
		ObjectSize:  int64(len(content)),
		ContentType: contentType,
	})
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
