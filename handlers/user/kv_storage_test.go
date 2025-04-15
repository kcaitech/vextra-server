package handlers

import (
	"testing"

	"kcaitech.com/kcserver/utils/sliceutil"
)

func TestAllowedKeyList(t *testing.T) {
	key := "Preferences"
	filted := sliceutil.FilterT(func(code string) bool {
		return key == code
	}, AllowedKeyList...)

	if len(filted) != 0 {
		t.Fatalf("%d", len(filted))
	}
}
