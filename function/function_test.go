package function

import (
	"testing"

	"github.com/johnny-morrice/godless/internal/testutil"
)

func TestStandardFunctions(t *testing.T) {
	functions := StandardFunctions()

	_, err := functions.GetFunction(StrEq{}.FuncName())
	testutil.AssertNil(t, err)

	_, err = functions.GetFunction("Foo")
	testutil.AssertNonNil(t, err)
}

func TestFunctionNamespace(t *testing.T) {
	functions := MakeFunctionNamespace()

	err := functions.PutFunction(StrEq{})
	testutil.AssertNil(t, err)

	_, err = functions.GetFunction(StrEq{}.FuncName())
	testutil.AssertNil(t, err)

	_, err = functions.GetFunction("Foo")
	testutil.AssertNonNil(t, err)
}
