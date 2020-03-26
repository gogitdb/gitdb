package gitdb

import (
	"testing"
)

func TestSetUser(t *testing.T) {
	setup()

	testDb.SetUser(NewUser("test", "tester@gitdb.io"))

	want := "test <tester@gitdb.io>"
	got := testDb.db().config.User.AuthorName()
	if want != got {
		t.Errorf("want: %s, got: %s", want, got)
	}

	got = testDb.db().config.User.String()
	if want != got {
		t.Errorf("want: %s, got: %s", want, got)
	}
}
