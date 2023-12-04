package sliceutil

import (
	"reflect"
)

func ConvertToInterfaceSlice(s any) []any {
	sValue := reflect.ValueOf(s)
	if !(sValue.IsValid() && (sValue.Kind() == reflect.Slice || sValue.Kind() == reflect.Array)) {
		return nil
	}
	length := sValue.Len()
	result := make([]any, length)
	for i := 0; i < length; i++ {
		result[i] = sValue.Index(i).Interface()
	}
	return result
}

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
