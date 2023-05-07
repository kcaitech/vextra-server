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

func Filter(fn func(item interface{}) bool, s ...interface{}) []interface{} {
	result := make([]interface{}, 0, len(s))
	for _, item := range s {
		if fn(item) {
			result = append(result, item)
		}
	}
	return result
}
