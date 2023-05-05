package minio

import (
	"context"
	"encoding/json"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"protodesign.cn/kcserver/utils/storage/base"
	"strconv"
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

var authOpMap = map[int]string{
	base.AuthOpGetObject:  "s3:GetObject",
	base.AuthOpPutObject:  "s3:PutObject",
	base.AuthOpDelObject:  "s3:DeleteObject",
	base.AuthOpListObject: "s3:ListBucket",
}

func (that *bucket) GenerateAccessKey(authPath string, authOp int, expires int) (*base.AccessKeyValue, error) {
	authPath = strings.TrimLeft(authPath, "/")
	authOpList := make([]string, 0, strconv.IntSize)
	authOpListDistinct := make(map[int]bool, strconv.IntSize)
	for i := 0; i < strconv.IntSize; i++ {
		authOpValue := 1 << i
		if authOpListDistinct[authOpValue] {
			continue
		}
		if authOp&(authOpValue) > 0 {
			authOpList = append(authOpList, authOpMap[authOpValue])
			authOpListDistinct[authOpValue] = true
		}
	}
	policy, err := json.Marshal(map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Action": authOpList,
				"Effect": "Allow",
				"Resource": []string{
					"arn:aws:s3:::" + that.config.BucketName + "/" + authPath,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	stsAssumeRole, err := credentials.NewSTSAssumeRole("http://"+that.client.config.Endpoint, credentials.STSAssumeRoleOptions{
		AccessKey:       that.client.config.StsAccessKeyID,
		SecretKey:       that.client.config.StsSecretAccessKey,
		Policy:          string(policy),
		Location:        "",
		DurationSeconds: expires,
	})
	if err != nil {
		return nil, err
	}
	v, err := stsAssumeRole.Get()
	if err != nil {
		return nil, err
	}
	value := base.AccessKeyValue{
		AccessKeyID:     v.AccessKeyID,
		SecretAccessKey: v.SecretAccessKey,
		SessionToken:    v.SessionToken,
		SignerType:      int(v.SignerType),
	}
	return &value, err
}
