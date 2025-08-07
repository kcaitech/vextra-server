/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

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
