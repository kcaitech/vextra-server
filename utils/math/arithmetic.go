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
