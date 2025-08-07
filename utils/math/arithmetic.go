/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package math

// IntDivideCeil 向上取整
func IntDivideCeil(a, b int) int {
	quotient := a / b
	remainder := a % b
	if remainder != 0 {
		quotient++
	}
	return quotient
}

// Max 取任意多个数的最大值，当为0个时返回0
func Max[
	V int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64,
](values ...V) V {
	if len(values) == 0 {
		return V(0)
	}
	max := values[0]
	for _, value := range values[1:] {
		if value > max {
			max = value
		}
	}
	return max
}
