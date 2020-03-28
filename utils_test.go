package gitdb_test

import (
	"testing"

	"github.com/fobilow/gitdb"
)

func TestRandStr(t *testing.T) {
	n := 10
	str := gitdb.RandStr(n)
	if len(str) != n {
		t.Errorf("RandStr(%d) returned string of length %d", n, len(str))
	}
}
