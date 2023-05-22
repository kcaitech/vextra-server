package services

import (
	"log"
	"protodesign.cn/kcserver/utils/sliceutil"
	"testing"
)

func TestGenerateSelectArgs(t *testing.T) {
	result := GenerateSelectArgs(&AccessRecordAndFavoritesQueryResItem{}, "")
	result1 := sliceutil.MapT(func(item *SelectArgs) SelectArgs {
		return *item
	}, *result...)
	log.Println(result1)
}

func TestGenerateJoinArgs(t *testing.T) {
	result := make([]DocumentQueryResItem, 0)
	result1 := GenerateJoinArgs(&result, "document", ParamArgs{"#user_id": "document.user_id"})
	result2 := sliceutil.MapT(func(item *JoinArgs) JoinArgs {
		return *item
	}, *result1...)
	log.Println(result2)
}
