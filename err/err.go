// Package err provides a structured, wrappable error type for the ddgo domain layer.
//
// Intent: offer a single error type, [DdgoError], that pairs a classification (one of the
// sentinel errors such as [ErrNotFound] or [ErrValidation]) with a human-readable cause and
// optional key/value details, while staying fully compatible with the standard library
// [errors.Is], [errors.As] and [errors.AsType] functions.
//
// Objective: allow errors to be nested arbitrarily, producing a flattened message chain
// ("outer cause: inner cause: sentinel") and an aggregated list of details across the
// whole chain. The whole chain is always walked, so the root cause and its details are
// never dropped, no matter how deep the nesting is.
package err

import (
	"errors"
	"log/slog"
	"slices"
	"strings"
)

const causeSeparator = ": "

// Sentinel errors that classify a [DdgoError]. Match them with [errors.Is].
//
// Their messages are stable, machine-readable codes in snake_case (e.g. "not_found"),
// not human-readable sentences: they are meant to be mapped by callers to HTTP status
// codes, API error codes or metrics labels, and they always appear as the last fragment
// of [DdgoError.Error]. For that reason they deliberately do not follow the usual
// "pkg: description" convention, and their text MUST NOT change once published, since
// external consumers may depend on it.
var (
	// ErrNotFound indicates that a requested resource could not be found.
	ErrNotFound = errors.New("not_found")
	// ErrInternal indicates an unexpected internal failure.
	ErrInternal = errors.New("internal")
	// ErrValidation indicates that input or state failed validation.
	ErrValidation = errors.New("validation")
	// ErrAlreadyExists indicates a duplicate or conflicting resource.
	ErrAlreadyExists = errors.New("already_exists")
	// ErrDependency indicates a failure in an external dependency.
	ErrDependency = errors.New("dependency")
	// ErrInvalidData indicates that the provided data is invalid.
	ErrInvalidData = errors.New("invalid_data")
	// ErrUnknown indicates an error whose nature could not be determined.
	ErrUnknown = errors.New("unknown")
)

// DdgoError is a domain error that wraps an underlying error together with a cause message
// and optional contextual details.
//
// The wrapped error is usually one of the package sentinels (e.g. [ErrNotFound]) or
// another *DdgoError, forming a chain that can be inspected with [errors.Is] and
// [errors.As]. The zero value is usable and renders as an empty message.
type DdgoError struct {
	err     error
	cause   string
	details []string
}

// New creates a *DdgoError that wraps err with the given cause and details.
//
// Parameters:
//   - err: the wrapped error; typically a package sentinel or another *DdgoError. It may be
//     nil, in which case only cause is rendered.
//   - cause: a human-readable description of what went wrong at this level. An empty
//     cause is omitted from the rendered message.
//   - details: optional contextual entries (e.g. "user_id=42") attached to this level.
//     Details are written verbatim to logs by [ToSlogAttr] and [DdgoError.LogValue], so
//     they MUST NOT contain personal data (names, e-mails, documents) or secrets
//     (passwords, tokens). Prefer opaque identifiers such as "user_id=42" over
//     "user=joão".
//
// Returns a non-nil *DdgoError. The details slice is copied, so later changes to the
// caller's slice do not affect the error.
func New(err error, cause string, details ...string) *DdgoError {
	return &DdgoError{
		err:     err,
		cause:   cause,
		details: slices.Clone(details),
	}
}

// Error implements the error interface.
//
// It returns the causes of the whole chain joined by ": ", from the outermost level to
// the innermost error, e.g. "parent error: user not found: not_found". Empty causes are
// skipped, and a nil *DdgoError, whether as receiver or as a wrapped link, contributes
// nothing instead of panicking.
func (e *DdgoError) Error() string {
	if e == nil {
		return ""
	}
	return e.mapErrorCauses()
}

// Unwrap returns the wrapped error, enabling [errors.Is] and [errors.As] to traverse the
// chain. It returns nil when no error was wrapped or when the receiver is nil.
func (e *DdgoError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

func (e *DdgoError) mapErrorCauses() string {
	fragments := make([]string, 0, 2)

	if e.cause != "" {
		fragments = append(fragments, e.cause)
	}

	if e.err != nil {
		var inner string
		// Only the direct link is inspected: errors.AsType would walk the whole chain and
		// silently drop any wrapper sitting between two *DdgoError values.
		if ddErr, ok := e.err.(*DdgoError); ok && ddErr != nil { //nolint:errorlint // direct link only
			inner = ddErr.mapErrorCauses()
		} else {
			inner = e.err.Error()
		}

		if inner != "" {
			fragments = append(fragments, inner)
		}
	}

	return strings.Join(fragments, causeSeparator)
}

// Details returns the details of every *DdgoError in the error tree, in depth-first
// order from the outermost level inwards.
//
// The walk follows any wrapper in between, such as fmt.Errorf with %w and errors.Join,
// so details are not lost when foreign errors sit between two *DdgoError values. The
// returned slice is always newly allocated, so callers may modify it freely; it is
// empty, never nil, when no details exist. A nil receiver returns nil.
func (e *DdgoError) Details() []string {
	if e == nil {
		return nil
	}
	return collectDetails(e, []string{})
}

// collectDetails walks the error tree depth-first, following both Unwrap() error and
// Unwrap() []error (errors.Join), and appends the details of every *DdgoError it finds.
func collectDetails(err error, out []string) []string {
	if err == nil {
		return out
	}

	if ddErr, ok := err.(*DdgoError); ok { //nolint:errorlint // each node is visited explicitly
		if ddErr == nil {
			return out
		}
		out = append(out, ddErr.details...)
	}

	switch u := err.(type) { //nolint:errorlint // inspecting the node itself, not the chain
	case interface{ Unwrap() []error }:
		for _, child := range u.Unwrap() {
			out = collectDetails(child, out)
		}
	case interface{ Unwrap() error }:
		out = collectDetails(u.Unwrap(), out)
	}

	return out
}

// LogValue implements [slog.LogValuer], so a *DdgoError logged as an attribute value is
// rendered as a group.
//
// The group always contains "message" (the full chain from [DdgoError.Error]) and, when
// this level has details, "detail" with them joined by " | ". Only the details of this
// level are included; use [ToSlogAttr] to log the details of the whole chain. A nil
// receiver returns the zero [slog.Value].
func (e *DdgoError) LogValue() slog.Value {
	if e == nil {
		return slog.Value{}
	}

	attrs := []slog.Attr{
		slog.String("message", e.Error()),
	}

	if len(e.details) > 0 {
		attrs = append(attrs, slog.String("detail", strings.Join(e.details, separator)))
	}

	return slog.GroupValue(attrs...)
}
