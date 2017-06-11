package debug

import (
	"fmt"
	"reflect"
)

func AssertLenEquals(slice, other interface{}) {
	aV := reflect.ValueOf(slice)
	bV := reflect.ValueOf(other)

	if aV.Len() != bV.Len() {
		panic(fmt.Sprintf("Mismatch len: %v %v", slice, other))
	}
}
