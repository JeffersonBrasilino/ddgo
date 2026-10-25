// Legacy domain errors, superseded by package github.com/jeffersonbrasilino/ddgo/err.
//
// Intent: Standardize error reporting so all layers use known error types
// that can be mapped to HTTP status codes or other responses.
// Objective: Make it easier to detect and handle domain errors (e.g. not
// found, validation, dependency failure) without string matching.
//
// Every exported identifier in this file is deprecated and kept only for backward
// compatibility. New code must use err.New with one of the err sentinels (err.ErrNotFound,
// err.ErrValidation, ...) and match errors with errors.Is.

package ddgo

import "fmt"

type (
	abstractError struct {
		message  string
		previous error
	}
	// NotFoundError indicates a requested resource does not exist.
	//
	// Deprecated: Use err.New(err.ErrNotFound, message) from
	// github.com/jeffersonbrasilino/ddgo/err and match it with
	// errors.Is(e, err.ErrNotFound).
	NotFoundError struct {
		abstractError
	}
	// InternalError indicates an unexpected internal failure.
	//
	// Deprecated: Use err.New(err.ErrInternal, message) from
	// github.com/jeffersonbrasilino/ddgo/err and match it with
	// errors.Is(e, err.ErrInternal).
	InternalError struct {
		abstractError
	}
	// ValidationError indicates input or state failed validation.
	//
	// Deprecated: Use err.New(err.ErrValidation, message) from
	// github.com/jeffersonbrasilino/ddgo/err and match it with
	// errors.Is(e, err.ErrValidation).
	ValidationError struct {
		abstractError
	}
	// AlreadyExistsError indicates a duplicate or conflicting resource.
	//
	// Deprecated: Use err.New(err.ErrAlreadyExists, message) from
	// github.com/jeffersonbrasilino/ddgo/err and match it with
	// errors.Is(e, err.ErrAlreadyExists).
	AlreadyExistsError struct {
		abstractError
	}
	// DependencyError indicates a failure in an external dependency.
	//
	// Deprecated: Use err.New(err.ErrDependency, message) from
	// github.com/jeffersonbrasilino/ddgo/err and match it with
	// errors.Is(e, err.ErrDependency).
	DependencyError struct {
		abstractError
	}
	// InvalidDataError indicates the provided data is invalid.
	//
	// Deprecated: Use err.New(err.ErrInvalidData, message) from
	// github.com/jeffersonbrasilino/ddgo/err and match it with
	// errors.Is(e, err.ErrInvalidData).
	InvalidDataError struct {
		abstractError
	}
)

// NewValidationError creates an error for validation failures.
//
// Parameters: message — human-readable description.
// Returns: *ValidationError implementing error.
//
// Deprecated: Use err.New(err.ErrValidation, message) from
// github.com/jeffersonbrasilino/ddgo/err.
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		abstractError{
			message: message,
		},
	}
}

// NewDependencyError creates an error for dependency failures.
//
// Parameters: message — human-readable description.
// Returns: *DependencyError implementing error.
//
// Deprecated: Use err.New(err.ErrDependency, message) from
// github.com/jeffersonbrasilino/ddgo/err.
func NewDependencyError(message string) *DependencyError {
	return &DependencyError{
		abstractError{
			message: message,
		},
	}
}

// NewNotFoundError creates an error for missing resources.
//
// Parameters: message — human-readable description.
// Returns: *NotFoundError implementing error.
//
// Deprecated: Use err.New(err.ErrNotFound, message) from
// github.com/jeffersonbrasilino/ddgo/err.
func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{
		abstractError{
			message: message,
		},
	}
}

// NewInternalError creates an error for internal failures.
//
// Parameters: message — human-readable description.
// Returns: *InternalError implementing error.
//
// Deprecated: Use err.New(err.ErrInternal, message) from
// github.com/jeffersonbrasilino/ddgo/err.
func NewInternalError(message string) *InternalError {
	return &InternalError{
		abstractError{
			message: message,
		},
	}
}

// NewAlreadyExistsError creates an error for duplicate resources.
//
// Parameters: message — human-readable description.
// Returns: *AlreadyExistsError implementing error.
//
// Deprecated: Use err.New(err.ErrAlreadyExists, message) from
// github.com/jeffersonbrasilino/ddgo/err.
func NewAlreadyExistsError(message string) *AlreadyExistsError {
	return &AlreadyExistsError{
		abstractError{
			message: message,
		},
	}
}

// NewInvalidDataError creates an error for invalid data.
//
// Parameters: message — human-readable description.
// Returns: *InvalidDataError implementing error.
//
// Deprecated: Use err.New(err.ErrInvalidData, message) from
// github.com/jeffersonbrasilino/ddgo/err.
func NewInvalidDataError(message string) *InvalidDataError {
	return &InvalidDataError{
		abstractError{
			message: message,
		},
	}
}

// buildMessage returns message; if previous is non-nil, appends "; previous: ...".
func (e *abstractError) buildMessage(message string, previous error) string {
	err := message
	if previous != nil {
		err = fmt.Sprintf("%s; previous: %v", message, previous.Error())
	}
	return err
}

// SetPreviousError sets the wrapped cause and returns the receiver for chaining.
//
// Parameters: previous — the underlying error (may be nil).
// Returns: the receiver so callers can chain.
// Behavior: Error() will include "; previous: <previous.Error()>" when non-nil.
//
// Deprecated: Pass the cause as the first argument of err.New from
// github.com/jeffersonbrasilino/ddgo/err, e.g. err.New(previous, message). To keep a
// classification as well, wrap both: err.New(errors.Join(err.ErrNotFound, previous),
// message).
func (e *abstractError) SetPreviousError(previous error) *abstractError {
	e.previous = previous
	return e
}

// Error implements the error interface.
//
// Returns: the error message; if a previous error was set, it is appended.
func (e *abstractError) Error() string {
	return e.buildMessage(e.message, e.previous)
}
