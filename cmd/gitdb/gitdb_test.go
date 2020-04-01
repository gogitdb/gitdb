package main

import "testing"

func Test_embedUI(t *testing.T) {
	err := embedUI()
	if err != nil {
		t.Errorf("embedUI() failed: %s", err)
	}

}
