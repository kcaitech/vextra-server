package str

import (
	"log"
	"strings"
	"testing"
)

func TestRandom(t *testing.T) {
	codeList := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		codeList = append(codeList, GetRandomAlphaNumStr(8))
	}
	log.Println("\n" + strings.Join(codeList, "\n"))
}
