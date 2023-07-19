package storage

import (
	"errors"
	"protodesign.cn/kcserver/common/storage/config"
	"protodesign.cn/kcserver/utils/storage"
	"protodesign.cn/kcserver/utils/storage/base"
)

var Client base.Client
var Bucket base.Bucket
var FilesBucket base.Bucket

func Init(filePath string) error {
	conf := config.LoadConfig(filePath)

	var providerConf base.Config
	switch conf.Storage.Provider {
	case base.MINIO:
		providerConf = conf.Minio
	case base.S3:
		providerConf = conf.S3
	default:
		return errors.New("不支持的provider")
	}

	providerConf.ClientConfig.Provider = conf.Storage.Provider

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
