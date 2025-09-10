package errors

import (
	"errors"
	"fmt"
)

var (
	ErrFileNotFound         = errors.New("file not found")
	ErrInvalidFileFormat    = errors.New("invalid file format")
	ErrSchemaValidation     = errors.New("schema validation failed")
	ErrSyncWindowClosed     = errors.New("sync window is closed")
	ErrExternalAPITimeout   = errors.New("external API timeout")
	ErrExternalAPIError     = errors.New("external API error")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrInvalidGradeValue    = errors.New("invalid grade value")
)

type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s' with value '%v': %s",
		e.Field, e.Value, e.Message)
}

type RetryableError struct {
	Err     error
	Message string
}

func (e RetryableError) Error() string {
	return fmt.Sprintf("retryable error: %s - %s", e.Message, e.Err.Error())
}

func (e RetryableError) Unwrap() error {
	return e.Err
}

func NewRetryableError(err error, message string) error {
	return RetryableError{
		Err:     err,
		Message: message,
	}
}
