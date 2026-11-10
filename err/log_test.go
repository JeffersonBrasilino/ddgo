package err_test

import (
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/jeffersonbrasilino/ddgo/err"
)

func TestLog_ToSlogAttr(t *testing.T) {
	t.Run("returns slog attrs", func(t *testing.T) {
		t.Parallel()
		er := err.New(err.ErrNotFound, "user not found", "detail-1", "detail-2")
		got := err.ToSlogAttr(er)
		expected := "error=[message=user not found: not_found detail=detail-1 | detail-2]"
		if got.String() != expected {
			t.Errorf("expected %s, got %s", expected, got.String())
		}
	})

	t.Run("nil error", func(t *testing.T) {
		t.Parallel()
		got := err.ToSlogAttr(nil)
		expected := slog.Attr{}
		if !got.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, got)
		}
	})

	t.Run("raw error", func(t *testing.T) {
		t.Parallel()
		e := errors.New("raw error")
		got := err.ToSlogAttr(e)

		expected := slog.Attr{Key: "error", Value: slog.GroupValue(slog.Attr{
			Key:   "message",
			Value: slog.StringValue("raw error"),
		})}

		if !got.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, got)
		}
	})

	t.Run("raw error with break line", func(t *testing.T) {
		t.Parallel()
		e := errors.New("raw error \n test")
		got := err.ToSlogAttr(e)

		expected := slog.Attr{Key: "error", Value: slog.GroupValue(slog.Attr{
			Key:   "message",
			Value: slog.StringValue("raw error  |  test"),
		})}

		if !got.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, got)
		}
	})

	t.Run("ddgo error wrapped by fmt.Errorf keeps outer context and details", func(t *testing.T) {
		t.Parallel()
		e := fmt.Errorf("handler: %w", err.New(err.ErrNotFound, "user", "id=1"))
		got := err.ToSlogAttr(e).String()
		expected := "error=[message=handler: user: not_found detail=id=1]"
		if got != expected {
			t.Errorf("expected %s, got %s", expected, got)
		}
	})

	t.Run("ddgo error without details omits the detail attribute", func(t *testing.T) {
		t.Parallel()
		got := err.ToSlogAttr(err.New(err.ErrNotFound, "x")).String()
		expected := "error=[message=x: not_found]"
		if got != expected {
			t.Errorf("expected %s, got %s", expected, got)
		}
	})

	t.Run("line breaks are flattened in message and details", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			input    error
			expected string
		}{
			{
				name:     "crlf in message",
				input:    errors.New("line one\r\nline two"),
				expected: "error=[message=line one | line two]",
			},
			{
				name:     "lone cr in message",
				input:    errors.New("line one\rline two"),
				expected: "error=[message=line one | line two]",
			},
			{
				name:     "line breaks in details",
				input:    err.New(err.ErrNotFound, "x", "a\nb", "c\r\nd"),
				expected: "error=[message=x: not_found detail=a | b | c | d]",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				if got := err.ToSlogAttr(tt.input).String(); got != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, got)
				}
			})
		}
	})
}
