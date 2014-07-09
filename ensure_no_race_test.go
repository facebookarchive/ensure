// +build !race

package ensure_test

import (
	"testing"

	"github.com/facebookgo/ensure"
)

func indirect(f ensure.Fataler) {
	ensure.StringContains(f, "foo", "bar")
}

func TestIndirectStackTrace(t *testing.T) {
	var c capture
	indirect(&c)
	c.Equal(t, `        github.com/facebookgo/ensure/ensure_no_race_test.go:13 indirect
github.com/facebookgo/ensure/ensure_no_race_test.go:20 TestIndirectStackTrace
expected substring "bar" was not found in "foo"`)
}
