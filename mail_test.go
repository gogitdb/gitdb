package gitdb_test

import (
	"testing"
)

func TestMailGetMails(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	mails := testDb.GetMails()
	if len(mails) > 0 {
		t.Errorf("testDb.GetMails() should be 0")
	}
}
