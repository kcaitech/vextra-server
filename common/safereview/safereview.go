package safereview

import (
	"errors"

	"kcaitech.com/kcserver/common/safereview/ali"
	"kcaitech.com/kcserver/common/safereview/baidu"
	"kcaitech.com/kcserver/common/safereview/base"
	"kcaitech.com/kcserver/common/safereview/config"
)

var Client base.Client

func Init(conf *config.SafeReviewConf) error {
	// conf := config.LoadConfig(filePath)

	switch conf.Provider {
	case base.Ali:
		if err := ali.Init(conf); err != nil {
			return err
		}
		Client = ali.Client
	case base.Baidu:
		if err := baidu.Init(conf); err != nil {
			return err
		}
		Client = baidu.Client
	default:
		return errors.New("不支持的provider")
	}

	return nil
}
