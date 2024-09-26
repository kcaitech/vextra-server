package storage

import (
	"errors"
	"kcaitech.com/kcserver/utils/storage/base"
	"kcaitech.com/kcserver/utils/storage/minio"
	"kcaitech.com/kcserver/utils/storage/oss"
	"kcaitech.com/kcserver/utils/storage/s3"
)

func NewClient(config *base.ClientConfig) (base.Client, error) {
	switch config.Provider {
	case base.MINIO:
		return minio.NewClient(config)
	case base.S3:
		return s3.NewClient(config)
	case base.OSS:
		return oss.NewClient(config)
	default:
		return nil, errors.New("不支持的provider")
	}
}
