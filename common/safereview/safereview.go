package safereview

import (
	"errors"
	"protodesign.cn/kcserver/common/safereview/ali"
	"protodesign.cn/kcserver/common/safereview/base"
	"protodesign.cn/kcserver/common/safereview/config"
)

var Client base.Client

func Init(filePath string) error {
	conf := config.LoadConfig(filePath)

	switch conf.Provider {
	case base.Ali:
		if err := ali.Init(filePath); err != nil {
			return err
		}
		Client = ali.Client
	default:
		return errors.New("不支持的provider")
	}

	return nil
}
