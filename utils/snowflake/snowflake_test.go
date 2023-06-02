package snowflake

import (
	"log"
	"testing"
)

func TestRace(t *testing.T) {
	snowFlake, _ := NewSnowFlake(1)
	idList := make([]int64, 0, 40000)
	for i := 0; i < 40000; i++ {
		idList = append(idList, snowFlake.NextId()&16383)
	}
	log.Println(idList[16383-2 : 16383+2])
}
