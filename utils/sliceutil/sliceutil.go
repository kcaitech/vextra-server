/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

package sliceutil

func Filter(fn func(item any) bool, args ...any) []any {
	result := make([]any, 0, len(args))
	for _, item := range args {
		if fn(item) {
			result = append(result, item)
		}
	}
	return result
}

func FilterT[T any](fn func(item T) bool, args ...T) []T {
	result := make([]T, 0, len(args))
	for _, item := range args {
		if fn(item) {
			result = append(result, item)
		}
	}
	return result
}

func MapT[T any, V any](fn func(item T) V, args ...T) []V {
	result := make([]V, 0, len(args))
	for _, item := range args {
		result = append(result, fn(item))
	}
	return result
}

func ConvertToAnySlice[T any](s []T) []any {
	return MapT(func(cmd T) any {
		return cmd
	}, s...)
}

func Find[T any](fn func(item T) bool, args ...T) *T {
	for _, item := range args {
		if fn(item) {
			return &item
		}
	}
	return nil
}

func FindIndex[T any](fn func(item T) bool, args ...T) int {
	for i, item := range args {
		if fn(item) {
			return i
		}
	}
	return -1
}

func Exists[T any](fn func(item T) bool, args ...T) bool {
	for _, item := range args {
		if fn(item) {
			return true
		}
	}
	return false
}

func Unique[T any](getKeyFn func(item T) any, args ...T) []T {
	keys := map[any]struct{}{}
	result := make([]T, 0, len(args))
	if getKeyFn == nil {
		for _, item := range args {
			if _, ok := keys[item]; !ok {
				keys[item] = struct{}{}
				result = append(result, item)
			}
		}
	} else {
		for _, item := range args {
			key := getKeyFn(item)
			if _, ok := keys[key]; !ok {
				keys[key] = struct{}{}
				result = append(result, item)
			}
		}
	}
	return result
}

func ToAny(args ...any) []any {
	return args
}
