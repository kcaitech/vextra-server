package reflect

import (
	"log"
	"reflect"
	"testing"
)

type StructA struct {
	FieldA string `json:"field_a"`
}

func Test0(t *testing.T) {
	a := &StructA{
		FieldA: "a",
	}
	dataType := EnterPointer(reflect.TypeOf(a))
	field := dataType.Field(0)
	log.Println(field)
}
