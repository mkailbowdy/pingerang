package validator

import (
	"slices"
	"strings"
	"unicode/utf8"
	"net/url"

)

type Validator struct {
	FieldErrors map[string]string
}

// Return true if there are no errors. i.e. Is the value Valid?
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0
}

// Add to Validator's FieldErrors
func (v *Validator) AddFieldError(key, message string) {
	// If v.FieldErrors doesn't exist yet, initialize it.
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}
	// If there's no key value pair for the error in question, add it to the map
	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

func (v *Validator) NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func (v *Validator) MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

func (v *Validator) ValidUrl(userUrl url.URL) bool {
	if (userUrl.Scheme == "" || userUrl.Host == "") {
		return false
	} else if userUrl.Scheme != "http" && userUrl.Scheme != "https" {
		return false
	}
	return true
}
