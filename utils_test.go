package gitdb

import (
	"testing"
)

func TestRandStr(t *testing.T) {
	n := 10
	str := RandStr(n)
	if len(str) != n {
		t.Errorf("RandStr(%d) returned string of length %d", n, len(str))
	}
}
