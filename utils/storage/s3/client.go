package s3

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sts"
	"protodesign.cn/kcserver/utils/storage/base"
	"strconv"
	"strings"
)

type client struct {
	config *base.ClientConfig
	sess   *session.Session
	client *s3.S3
}

func NewClient(config *base.ClientConfig) (base.Client, error) {
	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(config.Endpoint),
		Region:           aws.String(config.Region),
		Credentials:      credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, ""),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	return &client{
		config: config,
		sess:   sess,
		client: s3.New(sess),
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
	uploader := s3manager.NewUploader(that.client.sess)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(that.config.BucketName),
		Key:         aws.String(putObjectInput.ObjectName),
		Body:        putObjectInput.Reader,
		ContentType: aws.String(putObjectInput.ContentType),
	})
	if err != nil {
		return nil, err
	}
	versionID := ""
	if result.VersionID != nil {
		versionID = *result.VersionID
	}
	return &base.UploadInfo{
		VersionID: versionID,
	}, nil
}

func (that *bucket) GetObjectInfo(objectName string) (*base.ObjectInfo, error) {
	result, err := that.client.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(that.config.BucketName),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return nil, err
	}
	return &base.ObjectInfo{
		VersionID: *result.VersionId,
	}, nil
}

func (that *bucket) GetObject(objectName string) ([]byte, error) {
	result, err := that.client.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(that.config.BucketName),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(result.Body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var authOpMap = map[int]string{
	base.AuthOpGetObject:  "s3:GetObject",
	base.AuthOpPutObject:  "s3:PutObject",
	base.AuthOpDelObject:  "s3:DeleteObject",
	base.AuthOpListObject: "s3:ListBucket",
}

func (that *bucket) GenerateAccessKey(authPath string, authOp int, expires int, roleArn string, roleSessionName string) (*base.AccessKeyValue, error) {
	authPath = strings.TrimLeft(authPath, "/")
	authOpList := make([]string, 0, strconv.IntSize)
	authOpListDistinct := make(map[int]struct{}, strconv.IntSize)
	for i := 0; i < strconv.IntSize; i++ {
		authOpValue := 1 << i
		if _, ok := authOpListDistinct[authOpValue]; ok {
			continue
		}
		if authOp&(authOpValue) > 0 {
			authOpList = append(authOpList, authOpMap[authOpValue])
			authOpListDistinct[authOpValue] = struct{}{}
		}
	}
	policy, err := json.Marshal(map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{
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
	stsSvc := sts.New(that.client.sess)
	result, err := stsSvc.AssumeRole(&sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn), //"arn:aws:iam::123456789012:role/demo"
		RoleSessionName: aws.String(roleSessionName),
		Policy:          aws.String(string(policy)),
		DurationSeconds: aws.Int64(int64(expires)),
	})
	if err != nil {
		return nil, err
	}
	value := base.AccessKeyValue{
		AccessKey:       *result.Credentials.AccessKeyId,
		SecretAccessKey: *result.Credentials.SecretAccessKey,
		SessionToken:    *result.Credentials.SessionToken,
	}
	return &value, err
}

func (that *bucket) CopyObject(srcPath string, destPath string) (*base.UploadInfo, error) {
	_, err := that.client.client.CopyObject(&s3.CopyObjectInput{
		CopySource: aws.String(that.config.BucketName + "/" + srcPath),
		Bucket:     aws.String(that.config.BucketName),
		Key:        aws.String(destPath),
	})
	if err != nil {
		return nil, err
	}
	return &base.UploadInfo{}, nil
}

func (that *bucket) CopyDirectory(srcDirPath string, destDirPath string) (*base.UploadInfo, error) {
	if srcDirPath == "" || srcDirPath == "/" || destDirPath == "" || destDirPath == "/" {
		return nil, errors.New("路径不能为空")
	}
	if err := that.client.client.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket:    aws.String(that.config.BucketName),
		Prefix:    aws.String(srcDirPath),
		Delimiter: nil,
	}, func(result *s3.ListObjectsV2Output, b bool) bool {
		for _, objectInfo := range result.Contents {
			_, _ = that.CopyObject(*objectInfo.Key, strings.Replace(*objectInfo.Key, srcDirPath, destDirPath, 1))
		}
		return true
	}); err != nil {
		return nil, err
	}
	return &base.UploadInfo{}, nil
}
