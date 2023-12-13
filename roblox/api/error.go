package api

import (
	"fmt"
	"strings"
)

// ErrorResponse is a representation of a Roblox web API error.
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

type errorsResponse struct {
	Errors []ErrorResponse `json:"errors,omitempty"`
}

func (err ErrorResponse) Error() string {
	return fmt.Sprintf("response code %d: %s", err.Code, err.Message)
}

func (errs errorsResponse) Error() string {
	s := make([]string, len(errs.Errors))
	for i, e := range errs.Errors {
		s[i] = e.Error()
	}
	return strings.Join(s, "; ")
}

func (errs errorsResponse) Unwrap() error {
	if len(errs.Errors) == 0 {
		return nil
	}
	return errs.Errors[0]
}
