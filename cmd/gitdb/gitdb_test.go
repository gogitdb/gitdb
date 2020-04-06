package main

import (
	"os"
	"testing"
)

func Test_embedUI(t *testing.T) {
	err := embedUI()
	if err != nil {
		t.Errorf("embedUI() failed: %s", err)
	}
	os.Remove("./ui_static.go")
}
