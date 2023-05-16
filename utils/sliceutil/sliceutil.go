package sliceutil

import "reflect"

func ConvertToInterfaceSlice(s interface{}) []interface{} {
	sValue := reflect.ValueOf(s)
	if !(sValue.IsValid() && (sValue.Kind() == reflect.Slice || sValue.Kind() == reflect.Array)) {
		return nil
	}
	length := sValue.Len()
	result := make([]interface{}, length)
	for i := 0; i < length; i++ {
		result[i] = sValue.Index(i).Interface()
	}
	return result
}

func Filter(fn func(item interface{}) bool, args ...interface{}) []interface{} {
	result := make([]interface{}, 0, len(args))
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
