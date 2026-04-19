package util

import "errors"

var (
	ErrNotImplemented      = errors.New("not implemented")
	ErrNotTested           = errors.New("not tested")
	ErrNameLength          = errors.New("name length is not supported")
	ErrNameInvalid         = errors.New("name contains invalid characters")
	ErrEpicUUID            = errors.New("at most one epic coudl be link to a task")
	ErrPath                = errors.New("path segment is empty or too long")
	ErrPathInvalid         = errors.New("path segment contains invalid characters")
	ErrDuplicateTask       = errors.New("task already linked to this epic")
	ErrStorageNotSupported = errors.New("storage type is not supported")
	ErrAlreadyDone         = errors.New("already marked as done")
	ErrNotFound            = errors.New("not found")
)
