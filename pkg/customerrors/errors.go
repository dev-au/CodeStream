package customerrors

import (
	"errors"
)

type customError struct {
	baseError error
	err       error
}

func (c *customError) Error() string {
	if c.err != nil {
		return c.err.Error()
	}
	return c.baseError.Error()
}

func (c *customError) Unwrap() error {
	return c.err
}

func (c *customError) Is(target error) bool {
	if target == nil {
		return false
	}

	var t *customError
	if errors.As(target, &t) {
		return c.baseError == t.baseError
	}

	return c.baseError == target
}

var (
	ErrDataNotFound     = errors.New("data_not_found_error")
	ErrSystem           = errors.New("system_error")
	ErrExternalService  = errors.New("external_service_error")
	ErrValidation       = errors.New("validation_error")
	ErrPermissionDenied = errors.New("permission_denied_error")
	ErrDataDuplicated   = errors.New("data_duplicated_error")
)

func wrap(currentErr, baseError error) error {
	return &customError{baseError: baseError, err: currentErr}
}

func WrapSystemError(err error) error           { return wrap(err, ErrSystem) }
func WrapDataNotFoundError(err error) error     { return wrap(err, ErrDataNotFound) }
func WrapValidationError(err error) error       { return wrap(err, ErrValidation) }
func WrapPermissionDeniedError(err error) error { return wrap(err, ErrPermissionDenied) }
func WrapExternalServiceError(err error) error  { return wrap(err, ErrExternalService) }
func WrapDataDuplicatedError(err error) error   { return wrap(err, ErrDataDuplicated) }

func IsSystemError(err error) bool {
	if errors.Is(err, ErrExternalService) {
		return false
	}

	if errors.Is(err, ErrValidation) {
		return false
	}

	if errors.Is(err, ErrPermissionDenied) {
		return false
	}

	if errors.Is(err, ErrDataNotFound) {
		return false
	}

	return true
}
