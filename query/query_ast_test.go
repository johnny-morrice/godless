package query

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/johnny-morrice/godless/internal/testutil"
	"github.com/johnny-morrice/godless/log"
)

func TestQuoteUnquote(t *testing.T) {
	config := &quick.Config{
		MaxCount: ATOM_CHECK_COUNT,
		Values:   genQuoteToken,
	}

	err := quick.Check(quoteUnquoteOk, config)

	if err != nil {
		t.Error(err)
	}
}

func quoteUnquoteOk(token string) bool {
	unquoted, err := unquote(token)
	quoted := quote(unquoted)

	if err != nil {
		panic(err)
	}

	same := quoted == token

	if !same {
		log.Debug("token: %s quoted: %s", token, quoted)
	}

	return same
}

func genQuoteToken(values []reflect.Value, rand *rand.Rand) {
	const MAX_LEN = 50
	text := testutil.RandPoint(rand, MAX_LEN)
	gen := reflect.ValueOf(text)
	values[0] = gen
}

var _plain []string = []string{
	"abc\\de",
	"\\n",
	"abc\"de",
	"abc\ade",
	"abc\bde",
	"abc\fde",
	"abc\nde",
	"abc\rde",
	"abc\tde",
	"abc\vde",
	// "d*4g`KmT!i9ZnY\\s901R4D:KC9(mZ7qBct5YFJs~fz0IxU^\\vM3Es;\\Y5Jl:",
}
var _escaped []string = []string{
	"abc\\\\de",
	"\\\\n",
	"abc\\\"de",
	"abc\\ade",
	"abc\\bde",
	"abc\\fde",
	"abc\\nde",
	"abc\\rde",
	"abc\\tde",
	"abc\\vde",
	// "d*4g`KmT!i9ZnY\\\\s901R4D:KC9(mZ7qBct5YFJs~fz0IxU^\\\\vM3Es;\\\\Y5Jl:",
}

func TestQuote(t *testing.T) {
	for i, input := range _plain {
		expected := _escaped[i]
		actual := quote(input)

		if expected != actual {
			t.Error("Expected", expected, "but received", actual)
		}
	}
}

func TestUnquote(t *testing.T) {
	for i, input := range _escaped {
		expected := _plain[i]
		actual, err := unquote(input)

		if err != nil {
			panic(err)
		}

		if expected != actual {
			t.Error("Expected", expected, "but received", actual)
		}
	}
}

const ATOM_CHECK_COUNT = 100
