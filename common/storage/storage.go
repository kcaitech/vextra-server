package storage

import (
	"errors"

	"kcaitech.com/kcserver/common/storage/config"
	"kcaitech.com/kcserver/utils/storage"
	"kcaitech.com/kcserver/utils/storage/base"
)

var Client base.Client
var Bucket base.Bucket
var FilesBucket base.Bucket

func Init(conf *config.StorageConf) error {
	// conf := config.LoadConfig(filePath)

	var providerConf base.Config
	switch conf.Provider {
	case base.MINIO:
		providerConf = conf.Minio
	case base.S3:
		providerConf = conf.S3
	case base.OSS:
		providerConf = conf.Oss
	default:
		return errors.New("不支持的provider")
	}

	providerConf.ClientConfig.Provider = conf.Provider

	var err error
	if Client, err = storage.NewClient(&providerConf.ClientConfig); err != nil {
		return err
	}
	Bucket = Client.NewBucket(&base.BucketConfig{
		BucketName: providerConf.BucketName,
	})
	FilesBucket = Client.NewBucket(&base.BucketConfig{
		BucketName: providerConf.FilesBucketName,
	})
	return nil
}
