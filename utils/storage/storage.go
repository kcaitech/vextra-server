package storage

import (
	"errors"
	"protodesign.cn/kcserver/utils/storage/base"
	"protodesign.cn/kcserver/utils/storage/minio"
	"protodesign.cn/kcserver/utils/storage/s3"
)

func NewClient(config *base.ClientConfig) (base.Client, error) {
	switch config.Provider {
	case base.MINIO:
		return minio.NewClient(config)
	case base.S3:
		return s3.NewClient(config)
	default:
		return nil, errors.New("不支持的provider")
	}
}
