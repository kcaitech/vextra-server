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

// IsString 判断是否为字符串
func IsString(value any) bool {
	_, ok := value.(string)
	return ok
}

func IntToString(intVal int64) string {
	return strconv.FormatInt(intVal, 10)
}
