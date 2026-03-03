package entity

import "errors"

var ErrInvalid = errors.New("invalid")
var ErrConflict = errors.New("already exists")
var ErrNotFound = errors.New("not found")
