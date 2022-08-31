package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestNoSurf(t *testing.T) {
	var myH myHandler
	h := NoSurf(&myH)

	switch v := h.(type) {
	case http.Handler:
	//do nothing
	default:
		t.Error(fmt.Sprintf("The type is not an http.Handler but a %T", v))
	}
}

func TestSessionLoad(t *testing.T) {
	var myH myHandler
	h := SessionLoad(&myH)

	switch v := h.(type) {
	case http.Handler:
	//do nothing
	default:
		t.Error(fmt.Sprintf("The type is not an http.Handler but a %T", v))
	}
}
