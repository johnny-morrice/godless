package godless

import (
	"fmt"
	"reflect"
)

const __DEBUG bool = true

func assertLenEquals(slice, other interface{}) {
	aV := reflect.ValueOf(slice)
	bV := reflect.ValueOf(other)

	if aV.Len() != bV.Len() {
		panic(fmt.Sprintf("Mismatch len: %v %v", slice, other))
	}
}
