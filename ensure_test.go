package ensure

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"
)

var log = os.Getenv("ENSURE_LOG") == "1"

type capture struct {
	bytes.Buffer
}

func (c *capture) Fatal(a ...interface{}) {
	fmt.Fprint(&c.Buffer, a...)
}

func (c *capture) Equal(t testing.TB, expected string) {
	helper(t).Helper()
	DeepEqual(t, c.String(), expected)
	if log && expected != "" {
		t.Log(expected)
	}
}

func (c *capture) Contains(t testing.TB, suffix string) {
	helper(t).Helper()
	StringContains(t, c.String(), suffix)
	if log && suffix != "" {
		t.Log(suffix)
	}
}

func (c *capture) Matches(t testing.TB, pattern string) {
	helper(t).Helper()
	re := regexp.MustCompile(pattern)
	s := c.String()
	True(t, re.MatchString(s), s, "does not match pattern", pattern)
}

func TestNilErr(t *testing.T) {
	var c capture
	e := errors.New("foo")
	Err(&c, e, nil)
	c.Equal(t, "unexpected error: foo")
}

func TestMatchingError(t *testing.T) {
	var c capture
	e := errors.New("foo")
	Err(&c, e, regexp.MustCompile("bar"))
	c.Equal(t, "expected error: \"bar\" but got \"foo\"")
}

type typ struct {
	Answer int
}

func TestExtras(t *testing.T) {
	var c capture
	e := errors.New("foo")
	Err(
		&c,
		e,
		nil,
		map[string]int{"answer": 42},
		"baz",
		43,
		44.45,
		typ{Answer: 46},
	)
	c.Equal(t, `unexpected error: foo
(map[string]int) (len=1) {
 (string) (len=6) "answer": (int) 42
}
(string) (len=3) "baz"
(int) 43
(float64) 44.45
(ensure.typ) {
 Answer: (int) 46
}`)
}

func TestDeepEqualStruct(t *testing.T) {
	var c capture
	actual := typ{Answer: 41}
	expected := typ{Answer: 42}
	DeepEqual(&c, actual, expected)
	c.Equal(t, `expected these to be equal:
ACTUAL:
(ensure.typ) {
 Answer: (int) 41
}

EXPECTED:
(ensure.typ) {
 Answer: (int) 42
}`)
}

func TestDeepEqualString(t *testing.T) {
	var c capture
	DeepEqual(&c, "foo", "bar")
	c.Equal(t, `expected these to be equal:
ACTUAL:
(string) (len=3) "foo"

EXPECTED:
(string) (len=3) "bar"`)
}

func TestNotDeepEqualStruct(t *testing.T) {
	var c capture
	v := typ{Answer: 42}
	NotDeepEqual(&c, v, v)
	c.Equal(t, `expected two different values, but got the same:
(ensure.typ) {
 Answer: (int) 42
}`)
}

func TestUnexpectedNilErr(t *testing.T) {
	var c capture
	Err(&c, nil, regexp.MustCompile("bar"))
	c.Equal(t, "expected error: \"bar\" but got a nil error")
}

func TestNilString(t *testing.T) {
	var c capture
	Nil(&c, "foo")
	c.Equal(t, "expected nil value but got: (string) (len=3) \"foo\"")
}

func TestNilInt(t *testing.T) {
	var c capture
	Nil(&c, 1)
	c.Equal(t, "expected nil value but got: (int) 1")
}

func TestNilStruct(t *testing.T) {
	var c capture
	Nil(&c, typ{})
	c.Equal(t, `expected nil value but got:
(ensure.typ) {
 Answer: (int) 0
}`)
}

func TestNonNil(t *testing.T) {
	var c capture
	NotNil(&c, nil)
	c.Equal(t, `expected a value but got nil`)
}

func TestStringContains(t *testing.T) {
	var c capture
	StringContains(&c, "foo", "bar")
	c.Equal(t, "expected substring \"bar\" was not found in \"foo\"")
}

func TestStringDoesNotContain(t *testing.T) {
	var c capture
	StringDoesNotContain(&c, "foo", "o")
	c.Equal(t, "substring \"o\" was not supposed to be found in \"foo\"")
	if log {
		t.Log("foo")
	}
}

func TestExpectedNilErr(t *testing.T) {
	var c capture
	Err(&c, nil, nil)
	c.Equal(t, "")
}

func TestNilErrUsingNil(t *testing.T) {
	var c capture
	e := errors.New("foo")
	Nil(&c, e)
	c.Equal(t, "unexpected error: foo")
}

func TestTrue(t *testing.T) {
	var c capture
	True(&c, false)
	c.Equal(t, `expected true but got false`)
}

func TestSameElementsIntAndInterface(t *testing.T) {
	SameElements(t, []int{1, 2}, []interface{}{2, 1})
}

func TestSameElementsLengthDifference(t *testing.T) {
	var c capture
	SameElements(&c, []int{1, 2}, []interface{}{1})
	c.Equal(t, `expected same elements but found slices of different lengths:
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
	SameElements(&c, []int{1, 2}, []interface{}{1, 1})
	c.Equal(t, `missing expected element:
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

func TestFalse(t *testing.T) {
	var c capture
	False(t, false)
	False(&c, true)
	c.Equal(t, `expected false but got true`)
}

func TestPanicDeepEqualNil(t *testing.T) {
	defer PanicDeepEqual(t, "can't pass nil to ensure.PanicDeepEqual")
	PanicDeepEqual(t, nil)
}

func TestPanicDeepEqualSuccess(t *testing.T) {
	defer PanicDeepEqual(t, 1)
	panic(1)
}

func TestPanicDeepEqualFailure(t *testing.T) {
	var c capture
	func() {
		defer PanicDeepEqual(&c, 1)
		panic(2)
	}()
	c.Contains(t, `expected these to be equal:
ACTUAL:
(int) 2

EXPECTED:
(int) 1`)
}

func TestMultiLineStringContains(t *testing.T) {
	var c capture
	StringContains(&c, "foo\nbaz", "bar")
	c.Equal(t, `expected substring was not found:
EXPECTED SUBSTRING:
bar
ACTUAL:
foo
baz`)
}
