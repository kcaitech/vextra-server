package storage

import (
	"errors"
	"protodesign.cn/kcserver/utils/storage/base"
	"protodesign.cn/kcserver/utils/storage/minio"
)

func NewClient(config *base.ClientConfig) (base.Client, error) {
	switch config.Provider {
	case base.MINIO:
		return minio.NewClient(config)
	default:
		return nil, errors.New("不支持的provider")
	}
}
