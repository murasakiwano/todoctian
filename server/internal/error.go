package internal

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound      = errors.New("Resource not found.")
	ErrAlreadyExists = errors.New("Resource already exists.")
)

type RepositoryError struct {
	Err      error
	Resource string
}

func (e RepositoryError) Error() string {
	return fmt.Sprintf("%s not found", e.Resource)
}

func (e RepositoryError) Unwrap() error {
	return e.Err
}

func NewNotFoundError(resource string) RepositoryError {
	return RepositoryError{Resource: resource, Err: ErrNotFound}
}

func NewAlreadyExistsError(resource string) RepositoryError {
	return RepositoryError{Resource: resource, Err: ErrAlreadyExists}
}

type AlreadyExistsError struct {
	Err      error
	Resource string
}
