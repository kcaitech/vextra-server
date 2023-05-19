package services

import (
	"log"
	"protodesign.cn/kcserver/utils/sliceutil"
	"testing"
)

func Test0(t *testing.T) {
	result := GenerateSelectArgs(&AccessRecordAndFavoritesQueryResItem{}, "")
	result1 := sliceutil.MapT(func(item *SelectArgs) SelectArgs {
		return *item
	}, *result...)
	log.Println(result1)
}
