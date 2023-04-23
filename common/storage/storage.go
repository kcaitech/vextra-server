package storage

import (
	"errors"
	"protodesign.cn/kcserver/common/storage/config"
	"protodesign.cn/kcserver/utils/storage"
	"protodesign.cn/kcserver/utils/storage/base"
)

var Client base.Client
var Bucket base.Bucket

func Init() (base.Bucket, error) {
	conf := config.LoadConfig()

	var providerConf base.Config
	switch conf.Storage.Provider {
	case base.MINIO:
		providerConf = conf.Minio
	default:
		return nil, errors.New("不支持的provider")
	}

	var err error
	Client, err = storage.NewClient(&base.ClientConfig{
		Provider:        conf.Storage.Provider,
		Endpoint:        providerConf.Endpoint,
		AccessKeyID:     providerConf.AccessKeyID,
		SecretAccessKey: providerConf.SecretAccessKey,
	})
	if err != nil {
		return nil, err
	}
	Bucket = Client.NewBucket(&base.BucketConfig{
		BucketName: providerConf.BucketName,
		Region:     providerConf.Region,
	})
	return Bucket, nil
}
