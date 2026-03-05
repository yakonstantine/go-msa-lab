package entity

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalid = errors.New("invalid")
var ErrAlreadyExists = errors.New("already exists")
var ErrNotFound = errors.New("not found")

type ValidationError struct {
	Message string
	Fields  map[string]error
}

func (e *ValidationError) Error() string {
	var msgs []string
	for field, err := range e.Fields {
		msgs = append(msgs, fmt.Sprintf("%s: %v", field, err))
	}

	var prefix = e.Message
	if prefix == "" {
		prefix = "validation failed"
	}
	return prefix + ": " + strings.Join(msgs, "; ")
}
