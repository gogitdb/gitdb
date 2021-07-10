package gitdb_test

import "testing"

func TestUploadNew(t *testing.T) {
	teardown := setup(t, getConfig())
	defer teardown(t)

	err := testDb.Upload().New("creds", "./README.md")
	if err != nil {
		t.Errorf("Upload.New() failed: %s", err)
	}

}
