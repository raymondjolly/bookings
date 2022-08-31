package main

import (
	"net/http"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	//Before the program exit the test function will run my test
	os.Exit(m.Run())
}

type myHandler struct {
}

func (mh *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
