package strutil

import (
	"strconv"
)

func ToInt(strVal string) (int, error) {
	intVal, err := strconv.Atoi(strVal)
	if err != nil {
		return 0, err
	}
	return intVal, nil
}

func DefaultToInt(strVal string, defaultVal int) (int, error) {
	intVal, err := strconv.Atoi(strVal)
	if err != nil {
		return defaultVal, err
	}
	return intVal, nil
}
