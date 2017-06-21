package testutil

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/johnny-morrice/godless/internal/constants"
	"github.com/johnny-morrice/godless/internal/util"
	"github.com/johnny-morrice/godless/log"
)

func RandKey(rand *rand.Rand, max int) string {
	return RandLetters(rand, max)
}

func RandPoint(rand *rand.Rand, max int) string {
	const MIN_POINT_LENGTH = 0
	const pointSyms = constants.ALPHABET + constants.DIGITS + constants.SYMBOLS
	const injectScale = 0.1
	point := RandStr(rand, pointSyms, MIN_POINT_LENGTH, max)

	if len(point) == 0 {
		return point
	}

	if rand.Float32() > 0.5 {
		position := rand.Intn(len(point))
		inject := RandEscape(rand)
		point = Insert(point, inject, position)
	}

	return point
}

func Insert(old, ins string, pos int) string {
	before := old[:pos]
	after := old[pos:]
	return before + ins + after
}

func RandEscape(rand *rand.Rand) string {
	const chars = "\\tnav"
	const MIN_CHARS = 1
	const CHARS_LIM = 2
	return "\\" + RandStr(rand, chars, MIN_CHARS, CHARS_LIM)
}

// Fudge to generate count of sample data.
func GenCount(rand *rand.Rand, size int, scale float32) int {
	return GenCountRange(rand, 0, size, scale)
}

// Fudge to generate count of sample data.
func GenCountRange(rand *rand.Rand, min, max int, scale float32) int {
	fudge := float32(1.0)
	mark := rand.Float32()
	if mark < 0.01 {
		fudge = 0
	} else if mark < 0.3 {
		fudge = 0.3
	} else if mark < 0.7 {
		fudge = 0.5
	} else if mark < 0.9 {
		fudge = 0.8
	}

	gen := int(fudge * float32(max) * scale)
	if gen < min {
		gen = min
	}
	return gen
}

func RandLetters(rand *rand.Rand, max int) string {
	return RandStr(rand, constants.ALPHABET, 0, max)
}

func RandLettersRange(rand *rand.Rand, min, max int) string {
	return RandStr(rand, constants.ALPHABET, min, max)
}

func RandStr(rand *rand.Rand, elements string, min, max int) string {
	count := rand.Intn(max - min)
	count += min
	parts := make([]string, count)

	for i := 0; i < count; i++ {
		index := rand.Intn(len(elements))
		b := elements[index]
		parts[i] = string([]byte{b})
	}

	return strings.Join(parts, "")
}

func Trim(err error) string {
	msg := err.Error()

	const elipses = "..."

	if len(msg) < __TRIM_LENGTH+len(elipses) {
		return msg
	} else {
		return msg[:__TRIM_LENGTH] + elipses
	}
}

func DebugLine(t *testing.T) {
	_, _, line, ok := runtime.Caller(__CALLER_DEPTH)

	if !ok {
		panic("DebugLine failed")
	}

	t.Log("Test failed at line", line)
}

func AssertNil(t *testing.T, x interface{}) {
	Assert(t, fmt.Sprintf("Expected nil value but received: %v", x), x == nil)
}

func AssertNonNil(t *testing.T, x interface{}) {
	Assert(t, fmt.Sprintf("Expected non nil value"), x != nil)
}

func AssertBytesEqual(t *testing.T, expected, actual []byte) {
	if len(expected) != len(actual) {
		t.Error("Expected bytes length", len(expected), "but received", len(actual))
		return
	}

	for i, e := range expected {
		a := actual[i]
		if e != a {
			t.Error("Byte difference at", i)
		}
	}
}

func AssertEncodingStable(t *testing.T, expected []byte, encoder func(io.Writer)) {
	buff := &bytes.Buffer{}
	for i := 0; i < ENCODE_REPEAT_COUNT; i++ {
		encoder(buff)

		actual := buff.Bytes()

		AssertBytesEqual(t, expected, actual)
		buff.Reset()
	}
}

func Assert(t *testing.T, message string, isOk bool) {
	if !isOk {
		t.Error(message)
	}
}

func AssertEquals(t *testing.T, message string, expected, actual interface{}) {
	same := reflect.DeepEqual(expected, actual)

	if !same {
		expectedType := reflect.TypeOf(expected)
		actualType := reflect.TypeOf(actual)
		t.Errorf("%s: expected '%v' (%v) but received '%v' (%v)", message, expected, expectedType, actual, actualType)
	}
}

func AssertVerboseErrorIsNil(t *testing.T, err error) {
	if err != nil {
		t.Error("Unexpected error:", Trim(err))
	}
}

func AssertLenEquals(t *testing.T, expected int, hasLen interface{}) {
	value := reflect.ValueOf(hasLen)
	actual := value.Len()

	if expected != actual {
		t.Errorf("Expected len %v but received %v", expected, actual)
	}
}

func LogDiff(old, new string) {
	oldParts := strings.Split(old, "")
	newParts := strings.Split(new, "")

	minSize := util.Imin(len(oldParts), len(newParts))

	for i := 0; i < minSize; i++ {
		oldChar := oldParts[i]
		newChar := newParts[i]

		if oldChar != newChar {
			fragmentStart := i - 10
			if fragmentStart < 0 {
				fragmentStart = 0
			}

			fragmentEnd := i + 100

			oldEnd := fragmentEnd
			if len(old) < fragmentEnd {
				oldEnd = len(old) - 1
			}

			newEnd := fragmentEnd
			if len(new) < fragmentEnd {
				newEnd = len(new) - 1
			}

			oldFragment := old[fragmentStart:oldEnd]
			newFragment := new[fragmentStart:newEnd]

			log.Error("First difference at %v", i)
			log.Error("Old was: '%v'", oldFragment)
			log.Error("New was: '%v'", newFragment)
			return
		}
	}

	log.Error("logdiff called but no difference found")
}

const __CALLER_DEPTH = 2
const __TRIM_LENGTH = 500
const ENCODE_REPEAT_COUNT = 20

const KEY_SYMS = constants.ALPHABET + constants.DIGITS

type randGen struct {
	rand *rand.Rand
	sync.Mutex
}

var __rand randGen

func Rand() *rand.Rand {
	__rand.Lock()
	if __rand.rand == nil {
		seed := time.Now().UnixNano()
		src := rand.NewSource(seed)
		__rand.rand = rand.New(src)
	}
	__rand.Unlock()

	return __rand.rand
}

// Logging on in test mode!
func init() {
	log.SetLevel(log.LOG_DEBUG)
}
