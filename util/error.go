package util

import "errors"

var ErrNotImplemented = errors.New("not implemented")
var ErrNotTested = errors.New("not tested")
var ErrName = errors.New("name is empty or too long")
