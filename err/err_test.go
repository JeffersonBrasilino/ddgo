package err_test

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/jeffersonbrasilino/ddgo/err"
)

func TestError_Assertions(t *testing.T) {
	t.Run("is equal to sentinel", func(t *testing.T) {
		t.Parallel()
		target := err.New(err.ErrNotFound, "user not found")
		if !errors.Is(target, err.ErrNotFound) {
			t.Errorf("errors.Is() is not equal to err.ErrNotFound")
		}
	})

	t.Run("error message equals", func(t *testing.T) {
		t.Parallel()
		target := err.New(err.ErrNotFound, "user not found")
		if got := target.Error(); got != "user not found: not_found" {
			t.Errorf("Error() = %q, want %q", got, "user not found: not_found")
		}
	})

	t.Run("as ddgo error equals", func(t *testing.T) {
		t.Parallel()
		target := err.New(err.ErrNotFound, "user not found")
		if _, ok := errors.AsType[*err.DdgoError](target); !ok {
			t.Errorf("errors.AsType() is not equal to *err.DdgoError")
		}
	})

	t.Run("return only the error type when the error has no message", func(t *testing.T) {
		t.Parallel()
		target := err.New(err.ErrNotFound, "")
		if got := target.Error(); got != "not_found" {
			t.Errorf("Error() = %q, want %q", got, "not_found")
		}
	})

	t.Run("return a error with the error type and the message from nested error", func(t *testing.T) {
		t.Parallel()
		target := err.New(err.New(err.ErrNotFound, "user not found"), "parent user error")
		if got := target.Error(); got != "parent user error: user not found: not_found" {
			t.Errorf("Error() = %q, want %q", got, "parent user error: user not found: not_found")
		}
	})
}

func TestError_Details(t *testing.T) {
	t.Run("return details", func(t *testing.T) {
		t.Parallel()
		target := err.New(err.ErrNotFound, "user not found", "detail-1", "detail-2")
		got := target.Details()
		if len(got) != 2 {
			t.Errorf("Details() = %q, want %q", got, "user not found: not_found: detail-1 | detail-2")
		}

		if got[0] != "detail-1" {
			t.Errorf("Details() = %q, want %q", got[0], "detail-1")
		}

		if got[1] != "detail-2" {
			t.Errorf("Details() = %q, want %q", got[1], "detail-2")
		}
	})

	t.Run("deep chain keeps every detail, including the root one", func(t *testing.T) {
		t.Parallel()
		const levels = 1000
		target := err.New(err.ErrNotFound, "root", "detail-root")
		for range levels {
			target = err.New(target, "wrapper", "detail")
		}

		got := target.Details()
		if len(got) != levels+1 {
			t.Fatalf("Details() length = %d, want %d", len(got), levels+1)
		}
		if got[len(got)-1] != "detail-root" {
			t.Errorf("Details() last = %q, want %q", got[len(got)-1], "detail-root")
		}
	})

	t.Run("deep chain keeps the root cause and the sentinel", func(t *testing.T) {
		t.Parallel()
		const levels = 1000
		target := err.New(err.ErrNotFound, "root")
		for range levels {
			target = err.New(target, "wrapper")
		}

		want := strings.Repeat("wrapper: ", levels) + "root: not_found"
		if got := target.Error(); got != want {
			t.Errorf("Error() suffix = %q, want suffix %q", got[len(got)-30:], want[len(want)-30:])
		}
	})
}

func TestDdgoError_ForeignWrappers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		target          *err.DdgoError
		expectedMessage string
		expectedDetails []string
	}{
		{
			name: "fmt.Errorf between two ddgo errors keeps its context",
			target: err.New(
				fmt.Errorf("repo: %w", err.New(err.ErrNotFound, "user", "id=1")),
				"svc", "op=get",
			),
			expectedMessage: "svc: repo: user: not_found",
			expectedDetails: []string{"op=get", "id=1"},
		},
		{
			name: "errors.Join keeps every branch",
			target: err.New(
				errors.Join(
					err.New(err.ErrNotFound, "a", "da"),
					err.New(err.ErrValidation, "b", "db"),
				),
				"outer",
			),
			expectedMessage: "outer: a: not_found\nb: validation",
			expectedDetails: []string{"da", "db"},
		},
		{
			name:            "no details returns an empty non-nil slice",
			target:          err.New(err.ErrNotFound, "x"),
			expectedMessage: "x: not_found",
			expectedDetails: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.target.Error(); got != tt.expectedMessage {
				t.Errorf("Error() = %q, want %q", got, tt.expectedMessage)
			}

			got := tt.target.Details()
			if got == nil || !slices.Equal(got, tt.expectedDetails) {
				t.Errorf("Details() = %#v, want %#v", got, tt.expectedDetails)
			}
		})
	}
}

func TestError_LogValue(t *testing.T) {
	t.Run("return Log value", func(t *testing.T) {
		t.Parallel()

		target := err.New(err.ErrNotFound, "user not found", "detail-1", "detail-2")

		want := "[message=user not found: not_found detail=detail-1 | detail-2]"
		got := target.LogValue()
		if got.String() != want {
			t.Errorf("LogValue() = %v, want %v", got, want)
		}
	})
}

/*
go test -covermode=atomic -count=1 -race -coverprofile=coverage.out $(PACKAGES_TESTS)
	go tool cover -func=coverage.out
*/
