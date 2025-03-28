package safereview

import (
	"errors"
)

func NewSafeReview(conf *SafeReviewConf) (Client, error) {
	switch conf.Provider {
	case Ali:
		return NewAliClient(conf)
	case Baidu:
		return NewBaiduClient(conf)
	default:
		return nil, errors.New("不支持的provider")
	}
}

// var Client base.Client

// func Init(conf *config.SafeReviewConf) error {
// 	// conf := config.LoadConfig(filePath)

// 	switch conf.Provider {
// 	case base.Ali:
// 		if err := ali.Init(conf); err != nil {
// 			return err
// 		}
// 		Client = ali.Client
// 	case base.Baidu:
// 		if err := baidu.Init(conf); err != nil {
// 			return err
// 		}
// 		Client = baidu.Client
// 	default:
// 		return errors.New("不支持的provider")
// 	}

// 	return nil
// }
