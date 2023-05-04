package snowflake

import (
	"log"
	"testing"
)

func TestRace(t *testing.T) {
	snowFlake, _ := NewSnowFlake(1)
	var id int64
	for i := 0; i < 40000; i++ {
		id = snowFlake.NextId()
	}
	log.Println(id)
}
