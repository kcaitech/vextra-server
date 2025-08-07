/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package str

import (
	"strconv"
)

func ToInt(strVal string) (int64, error) {
	intVal, err := strconv.ParseInt(strVal, 10, 64)
	if err != nil {
		return 0, err
	}
	return intVal, nil
}

func DefaultToInt(strVal string, defaultVal int64) int64 {
	intVal, err := strconv.ParseInt(strVal, 10, 64)
	if err != nil {
		return defaultVal
	}
	return intVal
}

func ToBool(strVal string) (bool, error) {
	boolVal, err := strconv.ParseBool(strVal)
	if err != nil {
		return false, err
	}
	return boolVal, nil
}

func DefaultToBool(strVal string, defaultVal bool) bool {
	boolVal, err := strconv.ParseBool(strVal)
	if err != nil {
		return defaultVal
	}
	return boolVal
}

// IsString 判断是否为字符串
func IsString(value any) bool {
	_, ok := value.(string)
	return ok
}

func IntToString(intVal int64) string {
	return strconv.FormatInt(intVal, 10)
}
