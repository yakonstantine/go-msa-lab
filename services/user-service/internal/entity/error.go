package entity

import (
	"errors"
)

var ErrInvalid = errors.New("invalid")
var ErrAlreadyExists = errors.New("already exists")
var ErrNotFound = errors.New("not found")

type ValidationError struct {
	Message string
	Fields  map[string]error
}

func (e ValidationError) Error() string {
	var msg = e.Message
	if msg == "" {
		msg = "validation failed"
	}
	return msg
}
