package gitdb_test

import (
	"testing"
)

func TestModelString(t *testing.T) {
	m := getTestMessage()
	want := ""
	got := m.String()
	if got != want {
		t.Errorf("want: %s, got: %s", want, got)
	}
}

func TestModelGetValidationErrors(t *testing.T) {
	m := getTestMessage()
	m.Validate()

	errs := m.GetValidationErrors()
	if len(errs) > 0 {
		t.Errorf("m.GetValidationErrors() should be 0 not %d", len(errs))
	}
}
