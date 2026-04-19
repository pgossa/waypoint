package util

import (
	"regexp"
)

var (
	validSegment = regexp.MustCompile(`^[a-zA-Z0-9 ._-]+$`)
	validName    = regexp.MustCompile(`^[a-zA-Z0-9 ._/#@-]+$`)
)

func ValidateName(name string) error {
	if name == "" || len(name) > 64 {
		return ErrNameLength
	}
	if !validName.MatchString(name) {
		return ErrNameInvalid
	}
	return nil
}

func ValidatePath(path []string) error {
	if len(path) == 0 {
		return ErrPath
	}
	for _, segment := range path {
		if segment == "" || len(segment) > 64 {
			return ErrPath
		}
		if !validSegment.MatchString(segment) {
			return ErrPathInvalid
		}
	}
	return nil
}
