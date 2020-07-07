// Package ensure provides utilities for testing to ensure the
// given conditions are met and Fatal if they aren't satisified.
//
// The various functions here show a useful error message automatically
// including identifying source location. They additionally support arbitary
// arguments which will be printed using the spew library.
package ensure

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

// Fataler defines the minimal interface necessary to trigger a Fatal when a
// condition is hit. testing.T & testing.B satisfy this for example.
type Fataler interface {
	Fatal(a ...interface{})
}

type fatalerHelper interface {
	Fataler
	Helper()
}

type helperWrapper struct {
	Fataler
}

func (helperWrapper) Helper() {}

func helper(t Fataler) fatalerHelper {
	if t, ok := t.(fatalerHelper); ok {
		return t
	}
	return helperWrapper{t}
}

// cond represents a condition that wasn't satisfied, and is useful to generate
// log messages.
type cond struct {
	Fataler    Fataler
	Format     string
	FormatArgs []interface{}
	Extra      []interface{}
}

func (c cond) String() string {
	var b bytes.Buffer
	if c.Format != "" {
		fmt.Fprintf(&b, c.Format, c.FormatArgs...)
	}
	if len(c.Extra) != 0 {
		fmt.Fprint(&b, "\n")
		fmt.Fprint(&b, tsdump(c.Extra...))
	}
	return b.String()
}

// fatal triggers the fatal and logs the cond's message.
func fatal(c cond) {
	helper(c.Fataler).Helper()
	c.Fataler.Fatal(c.String())
}

// Err ensures the error satisfies the given regular expression.
func Err(t Fataler, err error, re *regexp.Regexp, a ...interface{}) {
	helper(t).Helper()
	if err == nil && re == nil {
		return
	}

	if err == nil && re != nil {
		fatal(cond{
			Fataler:    t,
			Format:     `expected error: "%s" but got a nil error`,
			FormatArgs: []interface{}{re},
			Extra:      a,
		})
		return
	}

	if err != nil && re == nil {
		fatal(cond{
			Fataler:    t,
			Format:     `unexpected error: %+v`,
			FormatArgs: []interface{}{err},
			Extra:      a,
		})
		return
	}

	if !re.MatchString(err.Error()) {
		fatal(cond{
			Fataler:    t,
			Format:     `expected error: "%s" but got "%+v"`,
			FormatArgs: []interface{}{re, err},
			Extra:      a,
		})
	}
}

// DeepEqual ensures actual and expected are equal. It does so using
// reflect.DeepEqual.
func DeepEqual(t Fataler, actual, expected interface{}, a ...interface{}) {
	helper(t).Helper()
	if !reflect.DeepEqual(actual, expected) {
		fatal(cond{
			Fataler:    t,
			Format:     "expected these to be equal:\nACTUAL:\n%s\nEXPECTED:\n%s",
			FormatArgs: []interface{}{spew.Sdump(actual), tsdump(expected)},
			Extra:      a,
		})
	}
}

// NotDeepEqual ensures actual and expected are not equal. It does so using
// reflect.DeepEqual.
func NotDeepEqual(t Fataler, actual, expected interface{}, a ...interface{}) {
	helper(t).Helper()
	if reflect.DeepEqual(actual, expected) {
		fatal(cond{
			Fataler:    t,
			Format:     "expected two different values, but got the same:\n%s",
			FormatArgs: []interface{}{tsdump(actual)},
			Extra:      a,
		})
	}
}

// Nil ensures v is nil.
func Nil(t Fataler, v interface{}, a ...interface{}) {
	helper(t).Helper()
	vs := tsdump(v)
	sp := " "
	if strings.Contains(vs[:len(vs)-1], "\n") {
		sp = "\n"
	}

	if v != nil {
		// Special case errors for prettier output.
		if _, ok := v.(error); ok {
			fatal(cond{
				Fataler:    t,
				Format:     `unexpected error: %+v`,
				FormatArgs: []interface{}{v},
				Extra:      a,
			})
		} else {
			fatal(cond{
				Fataler:    t,
				Format:     "expected nil value but got:%s%s",
				FormatArgs: []interface{}{sp, vs},
				Extra:      a,
			})
		}
	}
}

// NotNil ensures v is not nil.
func NotNil(t Fataler, v interface{}, a ...interface{}) {
	helper(t).Helper()
	if v == nil {
		fatal(cond{
			Fataler: t,
			Format:  "expected a value but got nil",
			Extra:   a,
		})
	}
}

// True ensures v is true.
func True(t Fataler, v bool, a ...interface{}) {
	helper(t).Helper()
	if !v {
		fatal(cond{
			Fataler: t,
			Format:  "expected true but got false",
			Extra:   a,
		})
	}
}

// False ensures v is false.
func False(t Fataler, v bool, a ...interface{}) {
	helper(t).Helper()
	if v {
		fatal(cond{
			Fataler: t,
			Format:  "expected false but got true",
			Extra:   a,
		})
	}
}

// StringContains ensures string s contains the string substr.
func StringContains(t Fataler, s, substr string, a ...interface{}) {
	helper(t).Helper()
	if !strings.Contains(s, substr) {
		format := `expected substring "%s" was not found in "%s"`

		// use multi line output if either string contains newlines
		if strings.Contains(s, "\n") || strings.Contains(substr, "\n") {
			format = "expected substring was not found:\nEXPECTED SUBSTRING:\n%s\nACTUAL:\n%s"
		}

		fatal(cond{
			Fataler:    t,
			Format:     format,
			FormatArgs: []interface{}{substr, s},
			Extra:      a,
		})
	}
}

// StringDoesNotContain ensures string s does not contain the string substr.
func StringDoesNotContain(t Fataler, s, substr string, a ...interface{}) {
	helper(t).Helper()
	if strings.Contains(s, substr) {
		fatal(cond{
			Fataler:    t,
			Format:     `substring "%s" was not supposed to be found in "%s"`,
			FormatArgs: []interface{}{substr, s},
			Extra:      a,
		})
	}
}

// SameElements ensures the two given slices contain the same elements,
// ignoring the order. It uses DeepEqual for element comparison.
func SameElements(t Fataler, actual, expected interface{}, extra ...interface{}) {
	helper(t).Helper()
	actualSlice := toInterfaceSlice(actual)
	expectedSlice := toInterfaceSlice(expected)
	if len(actualSlice) != len(expectedSlice) {
		fatal(cond{
			Fataler:    t,
			Format:     "expected same elements but found slices of different lengths:\nACTUAL:\n%s\nEXPECTED\n%s",
			FormatArgs: []interface{}{tsdump(actual), tsdump(expected)},
			Extra:      extra,
		})
	}

	used := map[int]bool{}
outer:
	for _, a := range expectedSlice {
		for i, b := range actualSlice {
			if !used[i] && reflect.DeepEqual(a, b) {
				used[i] = true
				continue outer
			}
		}
		fatal(cond{
			Fataler:    t,
			Format:     "missing expected element:\nACTUAL:\n%s\nEXPECTED:\n%s\nMISSING ELEMENT\n%s",
			FormatArgs: []interface{}{tsdump(actual), tsdump(expected), tsdump(a)},
			Extra:      extra,
		})
	}
}

// PanicDeepEqual ensures a panic occurs and the recovered value is DeepEqual
// to the expected value.
func PanicDeepEqual(t Fataler, expected interface{}, a ...interface{}) {
	helper(t).Helper()
	if expected == nil {
		panic("can't pass nil to ensure.PanicDeepEqual")
	}
	actual := recover()
	if !reflect.DeepEqual(actual, expected) {
		fatal(cond{
			Fataler:    t,
			Format:     "expected these to be equal:\nACTUAL:\n%s\nEXPECTED:\n%s",
			FormatArgs: []interface{}{spew.Sdump(actual), tsdump(expected)},
			Extra:      a,
		})
	}
}

// makes any slice into an []interface{}
func toInterfaceSlice(v interface{}) []interface{} {
	rv := reflect.ValueOf(v)
	l := rv.Len()
	ret := make([]interface{}, l)
	for i := 0; i < l; i++ {
		ret[i] = rv.Index(i).Interface()
	}
	return ret
}

// tsdump is Sdump without the trailing newline.
func tsdump(a ...interface{}) string {
	return strings.TrimSpace(spew.Sdump(a...))
}
