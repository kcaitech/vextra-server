package radix_convert

import "errors"

var Default62RadixChars = []string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
	"k", "l", "m", "n", "o", "p", "q", "r", "s", "t",
	"u", "v", "w", "x", "y", "z", "A", "B", "C", "D",
	"E", "F", "G", "H", "I", "J", "K", "L", "M", "N",
	"O", "P", "Q", "R", "S", "T", "U", "V", "W", "X",
	"Y", "Z",
}

type RadixConvert struct {
	radix         int
	radixChars    []string
	radixCharsMap map[string]int
}

func NewRadixConvert(radix int, radixChars []string) (*RadixConvert, error) {
	if radix < 2 || radix > len(radixChars) {
		return nil, errors.New("radix的取值必须在2-RadixChars.length之间")
	}
	radixCharsMap := map[string]int{}
	for i, v := range radixChars {
		radixCharsMap[v] = i
	}
	return &RadixConvert{
		radix:         radix,
		radixChars:    radixChars,
		radixCharsMap: radixCharsMap,
	}, nil
}

// From 整数转radix进制字符串
func (rc *RadixConvert) From(num int64) string {
	if num == 0 {
		return rc.radixChars[0]
	} else if num < 0 {
		return "-" + rc.From(-num)
	}
	result := ""
	for num > 0 {
		result = rc.radixChars[num%int64(rc.radix)] + result
		num /= int64(rc.radix)
	}
	return result
}

// To radix进制字符串转整数
func (rc *RadixConvert) To(str string) int64 {
	if str[0] == '-' {
		return -rc.To(str[1:])
	}
	result := int64(0)
	for _, v := range str {
		result = result*int64(rc.radix) + int64(rc.radixCharsMap[string(v)])
	}
	return result
}
