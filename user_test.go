package gitdb_test

import (
	"testing"

	"github.com/fobilow/gitdb"
)

func TestNewUser(t *testing.T) {
	user := gitdb.NewUser("test", "tester@gitdb.io")
	want := "test <tester@gitdb.io>"
	got := user.AuthorName()
	if want != got {
		t.Errorf("want: %s, got: %s", want, got)
	}

	got = user.String()
	if want != got {
		t.Errorf("want: %s, got: %s", want, got)
	}
}

func TestSetUser(t *testing.T) {
	teardown := setup(t, nil)
	defer teardown(t)

	user := gitdb.NewUser("test", "tester@gitdb.io")
	if err := testDb.SetUser(user); err != nil {
		t.Errorf("testDb.SetUser failed: %s", err)
	}
}
