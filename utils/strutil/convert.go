package strutil

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
