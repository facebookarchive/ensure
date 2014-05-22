package ensure_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/facebookgo/ensure"
)

var log = os.Getenv("ENSURE_LOG") == "1"

type capture struct {
	bytes.Buffer
}

func (c *capture) Fatal(a ...interface{}) {
	fmt.Fprint(&c.Buffer, a...)
}

var equalPrefix = strings.Repeat("\b", 20)

func (c *capture) Equal(t testing.TB, expected string) {
	// trim the deleteSelf '\b' prefix
	actual := strings.TrimLeft(c.String(), "\b")
	ensure.DeepEqual(t, actual, expected)
	if log && expected != "" {
		t.Log(equalPrefix, expected)
	}
}

func TestNilErr(t *testing.T) {
	var c capture
	e := errors.New("foo")
	ensure.Err(&c, e, nil)
	c.Equal(t, "ensure_test.go:40: unexpected error: foo")
}

func TestMatchingError(t *testing.T) {
	var c capture
	e := errors.New("foo")
	ensure.Err(&c, e, regexp.MustCompile("bar"))
	c.Equal(t, "ensure_test.go:47: expected error: \"bar\" but got \"foo\"")
}

type typ struct {
	Answer int
}

func TestExtras(t *testing.T) {
	var c capture
	e := errors.New("foo")
	ensure.Err(
		&c,
		e,
		nil,
		map[string]int{"answer": 42},
		"baz",
		43,
		44.45,
		typ{Answer: 46},
	)
	c.Equal(t, `ensure_test.go:76: unexpected error: foo
(map[string]int) (len=1) {
 (string) (len=6) "answer": (int) 42
}
(string) (len=3) "baz"
(int) 43
(float64) 44.45
(ensure_test.typ) {
 Answer: (int) 46
}`)
}

func TestDeepEqualStruct(t *testing.T) {
	var c capture
	actual := typ{Answer: 41}
	expected := typ{Answer: 42}
	ensure.DeepEqual(&c, actual, expected)
	c.Equal(t, `ensure_test.go:93: expected these to be equal:
ACTUAL:
(ensure_test.typ) {
 Answer: (int) 41
}

EXPECTED:
(ensure_test.typ) {
 Answer: (int) 42
}`)
}

func TestDeepEqualString(t *testing.T) {
	var c capture
	ensure.DeepEqual(&c, "foo", "bar")
	c.Equal(t, `ensure_test.go:104: expected these to be equal:
ACTUAL:
(string) (len=3) "foo"

EXPECTED:
(string) (len=3) "bar"`)
}

func TestNotDeepEqualStruct(t *testing.T) {
	var c capture
	v := typ{Answer: 42}
	ensure.NotDeepEqual(&c, v, v)
	c.Equal(t, `ensure_test.go:114: expected two different values, but got the same:
(ensure_test.typ) {
 Answer: (int) 42
}`)
}

func TestSubsetStruct(t *testing.T) {
	var c capture
	ensure.Subset(&c, typ{}, typ{Answer: 42})
	c.Equal(t, `ensure_test.go:129: expected subset not found:
ACTUAL:
(ensure_test.typ) {
 Answer: (int) 0
}

EXPECTED SUBSET
(ensure_test.typ) {
 Answer: (int) 42
}`)
}

func TestUnexpectedNilErr(t *testing.T) {
	var c capture
	ensure.Err(&c, nil, regexp.MustCompile("bar"))
	c.Equal(t, "ensure_test.go:135: expected error: \"bar\" but got a nil error")
}

func TestNilString(t *testing.T) {
	var c capture
	ensure.Nil(&c, "foo")
	c.Equal(t, "ensure_test.go:141: expected nil value but got: (string) (len=3) \"foo\"")
}

func TestNilInt(t *testing.T) {
	var c capture
	ensure.Nil(&c, 1)
	c.Equal(t, "ensure_test.go:147: expected nil value but got: (int) 1")
}

func TestNilStruct(t *testing.T) {
	var c capture
	ensure.Nil(&c, typ{})
	c.Equal(t, `ensure_test.go:156: expected nil value but got:
(ensure_test.typ) {
 Answer: (int) 0
}`)
}

func TestNonNil(t *testing.T) {
	var c capture
	ensure.NotNil(&c, nil)
	c.Equal(t, `ensure_test.go:162: expected a value but got nil`)
}

func TestStringContains(t *testing.T) {
	var c capture
	ensure.StringContains(&c, "foo", "bar")
	c.Equal(t, "ensure_test.go:168: expected substring \"bar\" was not found in \"foo\"")
}

func TestStringDoesNotContain(t *testing.T) {
	var c capture
	ensure.StringDoesNotContain(&c, "foo", "o")
	c.Equal(t, "ensure_test.go:174: substring \"o\" was not supposed to be found in \"foo\"")
	if log {
		t.Log("foo")
	}
}

func TestExpectedNilErr(t *testing.T) {
	var c capture
	ensure.Err(&c, nil, nil)
	c.Equal(t, "")
}

func indirect(f ensure.Fataler) {
	ensure.StringContains(f, "foo", "bar")
}

func TestIndirectStackTrace(t *testing.T) {
	var c capture
	indirect(&c)
	c.Equal(t, `        github.com/facebookgo/ensure/ensure_test.go:188 indirect
github.com/facebookgo/ensure/ensure_test.go:195 TestIndirectStackTrace
expected substring "bar" was not found in "foo"`)
}

func TestNilErrUsingNil(t *testing.T) {
	var c capture
	e := errors.New("foo")
	ensure.Nil(&c, e)
	c.Equal(t, "ensure_test.go:202: unexpected error: foo")
}

func TestTrue(t *testing.T) {
	var c capture
	ensure.True(&c, false)
	c.Equal(t, `ensure_test.go:208: expected true but got false`)
}

func TestSameElementsIntAndInterface(t *testing.T) {
	ensure.SameElements(t, []int{1, 2}, []interface{}{2, 1})
}

func TestSameElementsLengthDifference(t *testing.T) {
	var c capture
	ensure.SameElements(&c, []int{1, 2}, []interface{}{1})
	c.Equal(t, `ensure_test.go:227: expected same elements but found slices of different lengths:
ACTUAL:
([]int) (len=2 cap=2) {
 (int) 1,
 (int) 2
}
EXPECTED
([]interface {}) (len=1 cap=1) {
 (int) 1
}`)
}

func TestSameElementsRepeated(t *testing.T) {
	var c capture
	ensure.SameElements(&c, []int{1, 2}, []interface{}{1, 1})
	c.Equal(t, `ensure_test.go:245: missing expected element:
ACTUAL:
([]int) (len=2 cap=2) {
 (int) 1,
 (int) 2
}
EXPECTED:
([]interface {}) (len=2 cap=2) {
 (int) 1,
 (int) 1
}
MISSING ELEMENT
(int) 1`)
}
