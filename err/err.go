package err

import (
	"errors"
	"log/slog"
	"strings"
)

const (
	causeSeparator = ": "
	maxDepth       = 32
)

var (
	NotFoundError = errors.New("not_found")
	// InternalError indicates an unexpected internal failure.
	InternalError = errors.New("internal")
	// ValidationError indicates input or state failed validation.
	ValidationError = errors.New("validation")
	// AlreadyExistsError indicates a duplicate or conflicting resource.
	AlreadyExistsError = errors.New("already_exists")
	// DependencyError indicates a failure in an external dependency.
	DependencyError = errors.New("dependency")
	// InvalidDataError indicates the provided data is invalid.
	InvalidDataError = errors.New("invalid_data")
	// UnknownError indicates an unknown error.
	UnknownError = errors.New("unknown")
)

type DdgoErr struct {
	err     error
	cause   string
	details []string
}

func New(err error, cause string, details ...string) *DdgoErr {
	return &DdgoErr{
		err:     err,
		cause:   cause,
		details: details,
	}
}

func (e *DdgoErr) Error() string {
	return e.mapErrorCauses(0)
}

func (e *DdgoErr) Unwrap() error {
	return e.err
}

func (e *DdgoErr) mapErrorCauses(depth int) string {
	if depth > maxDepth {
		return ""
	}

	fragments := make([]string, 0, 2)

	if e.cause != "" {
		fragments = append(fragments, e.cause)
	}

	switch err := e.err.(type) {
	case *DdgoErr:
		if internalCause := err.mapErrorCauses(depth + 1); internalCause != "" {
			fragments = append(fragments, internalCause)
		}
	case error:
		if err.Error() != "" {
			fragments = append(fragments, err.Error())
		}
	}

	return strings.Join(fragments, causeSeparator)
}

func (e *DdgoErr) Details() []string {
	return e.mapErrorDetails(0)
}

func (e *DdgoErr) mapErrorDetails(depth int) []string {
	if depth > maxDepth {
		return nil
	}

	details := make([]string, 0)

	if len(e.details) > 0 {
		details = append(details, e.details...)
	}

	if err, ok := e.err.(*DdgoErr); ok {
		details = append(details, err.mapErrorDetails(depth+1)...)
	}

	return details
}

func (e *DdgoErr) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("message", e.Error()),
	}

	if len(e.details) > 0 {
		attrs = append(attrs, slog.String("detail", strings.Join(e.details, " | ")))
	}

	return slog.GroupValue(attrs...)
}

