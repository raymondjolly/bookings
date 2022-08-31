package main

import (
	"bookings/internal/config"
	"fmt"
	"github.com/go-chi/chi/v5"
	"reflect"
	"testing"
)

func TestRoutes(t *testing.T) {
	var app config.AppConfig

	mux := routes(&app)
	switch v := mux.(type) {
	case *chi.Mux:
	//do nothing
	default:
		t.Error(fmt.Sprintf("returned type is not of type %T but a %T", reflect.TypeOf(mux), v))
	}
}
