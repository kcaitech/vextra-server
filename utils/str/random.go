package str

import (
	"math/rand"
)

const (
	AlphaCharset   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	NumericCharset = "0123456789"
)

func getRandomStrFromCharset(charset string, length int) string {
	//r := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		randomIndex := rand.Intn(len(charset))
		result[i] = charset[randomIndex]
	}
	return string(result)
}

func GetRandomAlphaNumStr(length int) string {
	return getRandomStrFromCharset(AlphaCharset+NumericCharset, length)
}

func GetRandomAlphaStr(length int) string {
	return getRandomStrFromCharset(AlphaCharset, length)
}

func GetRandomNumStr(length int) string {
	return getRandomStrFromCharset(NumericCharset, length)
}
