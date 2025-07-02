package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aws/aws-sdk-go/aws"
	awsCreds "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/minio/minio-go/v7"
	minioCreds "github.com/minio/minio-go/v7/pkg/credentials"
	"kcaitech.com/kcserver/scripts/migrate-20250702/config"
)

// StorageInterface 定义存储操作接口
type StorageInterface interface {
	GetObject(bucketName, objectName string) ([]byte, error)
	ListObjects(bucketName, prefix string) <-chan ObjectInfo
}

// ObjectInfo 对象信息
type ObjectInfo struct {
	Key  string
	Size int64
	Err  error
}

// MinioStorage Minio存储实现
type MinioStorage struct {
	client *minio.Client
}

func NewMinioStorage(config config.Config) (*MinioStorage, error) {
	creds := minioCreds.NewStaticV4(config.Source.Storage.Minio.AccessKey, config.Source.Storage.Minio.SecretKey, "")
	client, err := minio.New(config.Source.Storage.Minio.Endpoint, &minio.Options{
		Creds:  creds,
		Region: config.Source.Storage.Minio.Region,
		Secure: false,
	})
	if err != nil {
		return nil, err
	}
	return &MinioStorage{client: client}, nil
}

func (m *MinioStorage) GetObject(bucketName, objectName string) ([]byte, error) {
	sourceObject, err := m.client.GetObject(context.Background(), bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer sourceObject.Close()
	return io.ReadAll(sourceObject)
}

func (m *MinioStorage) ListObjects(bucketName, prefix string) <-chan ObjectInfo {
	ch := make(chan ObjectInfo)
	go func() {
		defer close(ch)
		objectCh := m.client.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: true,
		})
		for object := range objectCh {
			ch <- ObjectInfo{
				Key:  object.Key,
				Size: object.Size,
				Err:  object.Err,
			}
		}
	}()
	return ch
}

// OSSStorage OSS存储实现
type OSSStorage struct {
	client *oss.Client
}

func NewOSSStorage(config config.Config) (*OSSStorage, error) {
	client, err := oss.New(config.Source.Storage.OSS.Endpoint, config.Source.Storage.OSS.AccessKey, config.Source.Storage.OSS.SecretKey, oss.Region(config.Source.Storage.OSS.Region))
	if err != nil {
		return nil, err
	}
	return &OSSStorage{client: client}, nil
}

func (o *OSSStorage) GetObject(bucketName, objectName string) ([]byte, error) {
	bucket, err := o.client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}

	readCloser, err := bucket.GetObject(objectName)
	if err != nil {
		return nil, err
	}
	defer readCloser.Close()

	return io.ReadAll(readCloser)
}

func (o *OSSStorage) ListObjects(bucketName, prefix string) <-chan ObjectInfo {
	ch := make(chan ObjectInfo)
	go func() {
		defer close(ch)
		bucket, err := o.client.Bucket(bucketName)
		if err != nil {
			ch <- ObjectInfo{Err: err}
			return
		}

		result, err := bucket.ListObjectsV2(oss.Prefix(prefix))
		if err != nil {
			ch <- ObjectInfo{Err: err}
			return
		}

		for _, object := range result.Objects {
			ch <- ObjectInfo{
				Key:  object.Key,
				Size: object.Size,
				Err:  nil,
			}
		}
	}()
	return ch
}

// S3Storage S3存储实现
type S3Storage struct {
	client *s3.S3
}

func NewS3Storage(config config.Config) (*S3Storage, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(config.Source.Storage.S3.Region),
		Endpoint:         aws.String(config.Source.Storage.S3.Endpoint),
		Credentials:      awsCreds.NewStaticCredentials(config.Source.Storage.S3.AccessKey, config.Source.Storage.S3.SecretKey, ""),
		S3ForcePathStyle: aws.Bool(true), // 用于兼容MinIO等S3兼容存储
	})
	if err != nil {
		return nil, err
	}

	client := s3.New(sess)
	return &S3Storage{client: client}, nil
}

func (s *S3Storage) GetObject(bucketName, objectName string) ([]byte, error) {
	result, err := s.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

func (s *S3Storage) ListObjects(bucketName, prefix string) <-chan ObjectInfo {
	ch := make(chan ObjectInfo)
	go func() {
		defer close(ch)

		input := &s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
			Prefix: aws.String(prefix),
		}

		err := s.client.ListObjectsV2Pages(input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, object := range page.Contents {
				ch <- ObjectInfo{
					Key:  aws.StringValue(object.Key),
					Size: aws.Int64Value(object.Size),
					Err:  nil,
				}
			}
			return !lastPage
		})

		if err != nil {
			ch <- ObjectInfo{Err: err}
		}
	}()
	return ch
}

// CreateStorageClient 根据provider创建存储客户端
func CreateStorageClient(config config.Config) (StorageInterface, error) {
	switch config.Source.Storage.Provider {
	case "minio":
		return NewMinioStorage(config)
	case "oss":
		return NewOSSStorage(config)
	case "s3":
		return NewS3Storage(config)
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", config.Source.Storage.Provider)
	}
}
