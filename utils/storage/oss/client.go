package oss

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"net/http"
	"protodesign.cn/kcserver/utils/storage/base"
	"strconv"
	"strings"
)

type client struct {
	config *base.ClientConfig
	client *oss.Client
}

func NewClient(config *base.ClientConfig) (base.Client, error) {
	c, err := oss.New(config.Endpoint, config.AccessKeyID, config.SecretAccessKey, oss.Region(config.Region))
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
	bucket *oss.Bucket
}

func (that *client) NewBucket(config *base.BucketConfig) base.Bucket {
	instance := &bucket{
		config: config,
		client: that,
	}
	instance.bucket, _ = that.client.Bucket(config.BucketName)
	instance.That = instance
	return instance
}

func (that *bucket) GetConfig() *base.Config {
	return &base.Config{
		ClientConfig: *that.client.config,
		BucketConfig: *that.config,
	}
}

func (that *bucket) PutObject(putObjectInput *base.PutObjectInput) (*base.UploadInfo, error) {
	if putObjectInput.ContentType == "" {
		putObjectInput.ContentType = "application/octet-stream"
		splitRes := strings.Split(putObjectInput.ObjectName, ".")
		if len(splitRes) > 1 {
			switch splitRes[len(splitRes)-1] {
			case "json":
				putObjectInput.ContentType = "application/json"
			}
		}
	}
	var retHeader http.Header
	err := that.bucket.PutObject(
		strings.TrimLeft(putObjectInput.ObjectName, "/"),
		putObjectInput.Reader,
		oss.ContentType(putObjectInput.ContentType),
		oss.ContentLength(putObjectInput.ObjectSize),
		oss.GetResponseHeader(&retHeader),
	)
	if err != nil {
		return nil, err
	}
	return &base.UploadInfo{
		VersionID: retHeader.Get("x-oss-version-id"),
	}, nil
}

func (that *bucket) GetObjectInfo(objectName string) (*base.ObjectInfo, error) {
	meta, err := that.bucket.GetObjectMeta(objectName)
	if err != nil {
		return nil, err
	}
	return &base.ObjectInfo{
		VersionID: meta.Get("x-oss-version-id"),
	}, nil
}

func (that *bucket) GetObject(objectName string) ([]byte, error) {
	readCloser, err := that.bucket.GetObject(objectName)
	if err != nil {
		return nil, err
	}
	defer readCloser.Close()
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(readCloser)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var authOpMap = map[int][]string{
	base.AuthOpGetObject:  {"oss:GetObject", "oss:GetObjectAcl", "oss:GetObjectVersion", "oss:GetObjectVersionAcl"},
	base.AuthOpPutObject:  {"oss:PutObject", "oss:PutObjectAcl", "oss:PutObjectVersionAcl"},
	base.AuthOpDelObject:  {"oss:DeleteObject", "oss:DeleteObjectVersion"},
	base.AuthOpListObject: {"oss:ListObjects", "oss:ListObjectVersions"},
}

func (that *bucket) GenerateAccessKey(authPath string, authOp int, expires int, roleSessionName string) (*base.AccessKeyValue, error) {
	authPath = strings.TrimLeft(authPath, "/")
	authOpList := make([]string, 0, strconv.IntSize)
	authOpListDistinct := make(map[int]struct{}, strconv.IntSize)
	for i := 0; i < strconv.IntSize; i++ {
		authOpValue := 1 << i
		if _, ok := authOpListDistinct[authOpValue]; ok {
			continue
		}
		if authOp&(authOpValue) > 0 {
			for _, v := range authOpMap[authOpValue] {
				authOpList = append(authOpList, v)
			}
			authOpListDistinct[authOpValue] = struct{}{}
		}
	}
	policy, err := json.Marshal(map[string]any{
		"Version": "1",
		"Statement": []map[string]any{
			{
				"Action": authOpList,
				"Effect": "Allow",
				"Resource": []string{
					"acs:oss:*:*:" + that.config.BucketName + "/" + authPath,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	roleArn := "acs:ram::" + that.client.config.AccountId + ":role/" + that.client.config.RoleName
	stsClient, err := sts.NewClientWithAccessKey(that.client.config.Region, that.client.config.AccessKeyID, that.client.config.SecretAccessKey)
	if err != nil {
		return nil, err
	}
	request := sts.CreateAssumeRoleRequest()
	request.Scheme = "https"
	request.Domain = that.client.config.StsEndpoint
	request.RoleArn = roleArn
	request.RoleSessionName = roleSessionName
	request.Policy = string(policy)
	request.DurationSeconds = requests.NewInteger(expires)
	response, err := stsClient.AssumeRole(request)
	if err != nil {
		return nil, err
	}
	v := response.Credentials
	return &base.AccessKeyValue{
		AccessKey:       v.AccessKeyId,
		SecretAccessKey: v.AccessKeySecret,
		SessionToken:    v.SecurityToken,
		SignerType:      1,
	}, nil
}

func (that *bucket) CopyObject(srcPath string, destPath string) (*base.UploadInfo, error) {
	_, err := that.bucket.CopyObject(srcPath, destPath)
	if err != nil {
		return nil, err
	}
	return &base.UploadInfo{}, nil
}

func (that *bucket) CopyDirectory(srcDirPath string, destDirPath string) (*base.UploadInfo, error) {
	if srcDirPath == "" || srcDirPath == "/" || destDirPath == "" || destDirPath == "/" {
		return nil, errors.New("路径不能为空")
	}
	result, err := that.bucket.ListObjectsV2(oss.Prefix(srcDirPath))
	if err != nil {
		return nil, err
	}
	for _, objectInfo := range result.Objects {
		_, _ = that.CopyObject(objectInfo.Key, strings.Replace(objectInfo.Key, srcDirPath, destDirPath, 1))
	}
	return &base.UploadInfo{}, nil
}
