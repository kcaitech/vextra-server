package math

func IntDivideCeil(a, b int) int {
	quotient := a / b
	remainder := a % b
	if remainder != 0 {
		quotient++
	}
	return quotient
}
