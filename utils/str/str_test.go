package str

import (
	"log"
	"strings"
	"testing"
)

func TestRandom(t *testing.T) {
	l := 10
	codeList := make([]string, 0, l)
	for i := 0; i < l; i++ {
		codeList = append(codeList, "\""+GetRandomAlphaNumStr(8)+"\",")
	}
	log.Println("\n" + strings.Join(codeList, "\n"))
}
