package main

import (
	"testing"
)

func TestRun(t *testing.T) {
	var err error
	_, err = run()
	if err != nil {
		t.Error("Failed Run!")
	}
}
