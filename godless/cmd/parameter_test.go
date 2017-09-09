package cmd

import (
	"testing"
	"time"

	"github.com/johnny-morrice/godless/internal/testutil"
)

func TestParameters(t *testing.T) {
	params := &Parameters{}

	strWrite := params.String("hello")
	testutil.AssertEquals(t, "Unexpected string", "", *strWrite)
	*strWrite = "awesome"

	strRead := params.String("hello")
	testutil.AssertEquals(t, "Unexpected string", "awesome", *strRead)

	strSliceWrite := params.StringSlice("hello")
	testutil.AssertLenEquals(t, 0, *strSliceWrite)
	*strSliceWrite = []string{"awesome"}

	strSliceRead := params.StringSlice("hello")
	testutil.AssertLenEquals(t, 1, *strSliceRead)

	intWrite := params.Int("hello")
	testutil.AssertEquals(t, "Unexpected int", 0, *intWrite)
	*intWrite = 42

	intRead := params.Int("hello")
	testutil.AssertEquals(t, "Unexpected int", 42, *intRead)

	durWrite := params.Duration("hello")
	testutil.AssertEquals(t, "Unexpected duration", time.Duration(0), *durWrite)
	*durWrite = 42

	durRead := params.Duration("hello")
	testutil.AssertEquals(t, "Unexpected duration", time.Duration(42), *durRead)

	boolWrite := params.Bool("hello")
	testutil.AssertEquals(t, "Unexpected duration", false, *boolWrite)
	*boolWrite = true

	boolRead := params.Bool("hello")
	testutil.AssertEquals(t, "Unexpected duration", true, *boolRead)
}
