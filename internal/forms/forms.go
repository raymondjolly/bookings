package forms

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"net/url"
	"strings"
)

// Form creates a custom form struct, embeds a url.Values object
type Form struct {
	url.Values
	Errors errors
}

// New initializes a new Form
func New(data url.Values) *Form {
	return &Form{
		data,
		map[string][]string{},
	}
}

// Has checks if form field is in POST and not empty
func (f *Form) Has(field string) bool {
	x := f.Get(field)
	if x == "" {
		return false
	}
	return true
}

// Required takes variadic parameters to declare input fields as required for backend validation.
// This is needed in case a user turns off Javascript in the browser
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "this field cannot be null")
		}
	}
}

// MinLength takes as field, length, and a pointer to a http request
// and checks field for minimum input length
func (f *Form) MinLength(field string, length int) bool {
	x := f.Get(field)
	if len(x) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long", length))
		return false
	}
	return true
}

// Valid returns true if there are no errors, otherwise false
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// IsEmail checks for valid email address input
func (f *Form) IsEmail(field string) {
	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "invalid email address")
	}
}
