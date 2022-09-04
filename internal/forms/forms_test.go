package forms

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T) {
	r := httptest.NewRequest("POST", "/some_link", nil)
	form := New(r.PostForm)

	isValid := form.Valid()
	if !isValid {
		t.Error("got invalid when should be valid")
	}
}

func TestForm_Required(t *testing.T) {
	r := httptest.NewRequest("POST", "/some_link", nil)
	form := New(r.PostForm)

	form.Required("a", "b", "c")
	if form.Valid() {
		t.Error("form shows valid when required fields are missing")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "a")
	postedData.Add("c", "a")

	r, _ = http.NewRequest("POST", "/some_link", nil)
	r.PostForm = postedData
	form = New(r.PostForm)
	form.Required("a", "b", "c")
	if !form.Valid() {
		t.Error("shows does not have required fields when the data is indeed valid")
	}
}

func TestForm_Has(t *testing.T) {
	r := httptest.NewRequest("POST", "/some_link", nil)
	form := New(r.PostForm)
	has := form.Has("some-value")
	if has {
		t.Error("form shows valid when value in form does not exist")
	}
	postedData := url.Values{}
	postedData.Add("a", "a")
	form = New(postedData)
	has = form.Has("a")
	if !has {
		t.Error("shows does not have value when the values are indeed there")
	}
}

func TestForm_MinLength(t *testing.T) {
	r := httptest.NewRequest("POST", "/some_link", nil)
	form := New(r.PostForm)
	form.MinLength("field", 7)
	if form.Valid() {
		t.Error("form shows a minimum length when one does not exist")
	}

	isError := form.Errors.Get("field")
	if isError == "" {
		t.Error("should have minLength error but did not get one")
	}

	postedValues := url.Values{}
	postedValues.Add("field", "some value")
	form = New(postedValues)
	form.MinLength("field", 100)
	if form.Valid() {
		t.Error("shows minlength of 100 met when data is shorter")
	}

	postedValues = url.Values{}
	postedValues.Add("field", "Some very long string")
	form.MinLength("field", 5)
	form = New(postedValues)
	if !form.Valid() {
		t.Error("shows min length of 5 when values are longer")
	}

	isError = form.Errors.Get("field")
	if isError != "" {
		t.Error("show not have an error but got one")
	}

}

func TestForm_IsEmail(t *testing.T) {
	postedValue := url.Values{}
	form := New(postedValue)
	form.IsEmail("x")
	if form.Valid() {
		t.Error("form shows a valid email when one does not exist")
	}
	postedValue = url.Values{}
	postedValue.Add("some_field", "test@email.com")
	form = New(postedValue)
	form.IsEmail("some_field")
	if !form.Valid() {
		t.Error("form shows invalid when value is indeed valid")
	}

	postedValue = url.Values{}
	postedValue.Add("another_field", "@test.go_")
	form = New(postedValue)
	form.IsEmail("another_email")
	if form.Valid() {
		t.Error("form shows a valid value when the email is invalid")
	}
}
