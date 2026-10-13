package err_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/jeffersonbrasilino/ddgo/err"
)

func TestError_Assertions(t *testing.T) {
	t.Run("is equal to sentinel", func(t *testing.T) {
		t.Parallel()
		target := err.New(err.NotFoundError, "user not found")
		if !errors.Is(target, err.NotFoundError) {
			t.Errorf("errors.Is() is not equal to err.NotFoundError")
		}
	})

	t.Run("error message equals", func(t *testing.T) {
		t.Parallel()
		target := err.New(err.NotFoundError, "user not found")
		if got := target.Error(); got != "user not found: not_found" {
			t.Errorf("Error() = %q, want %q", got, "user not found: not_found")
		}
	})

	t.Run("as ddgo error equals", func(t *testing.T) {
		t.Parallel()
		target := err.New(err.NotFoundError, "user not found")
		if _, ok := errors.AsType[*err.DdgoErr](target); !ok {
			t.Errorf("errors.AsType() is not equal to *err.DdgoErr")
		}
	})

	t.Run("return only the error type when the error has no message", func(t *testing.T) {
		t.Parallel()
		target := err.New(err.NotFoundError, "")
		if got := target.Error(); got != "not_found" {
			t.Errorf("Error() = %q, want %q", got, "not_found")
		}
	})

	t.Run("return a error with the error type and the message from nested error", func(t *testing.T) {
		t.Parallel()
		target := err.New(err.New(err.NotFoundError, "user not found"), "parent user error")
		if got := target.Error(); got != "parent user error: user not found: not_found" {
			t.Errorf("Error() = %q, want %q", got, "parent user error: user not found: not_found")
		}
	})
}
func TestError_Details(t *testing.T) {
	t.Run("return details", func(t *testing.T) {
		t.Parallel()
		target := err.New(err.NotFoundError, "user not found", "detail-1", "detail-2")
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

	t.Run("max depth of details", func(t *testing.T) {
		t.Parallel()
		var target = err.New(nil, "root", "detail-root")
		for i := 0; i < 34; i++ {
			target = err.New(target, "wrapper", "detail")
		}
		if got := target.Details(); len(got) != 33 {
			t.Errorf("Details() length = %d, want 33", len(got))
		}
	})

	t.Run("max depth of errors", func(t *testing.T) {
		t.Parallel()
		var target error = errors.New("root error")
		var expectedCauses []string
		for i := 0; i < 34; i++ {
			target = err.New(target, "wrapper")
			if i < 33 {
				expectedCauses = append(expectedCauses, "wrapper")
			}
		}
		want := strings.Join(expectedCauses, ": ")
		if got := target.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})
}
func TestError_LogValue(t *testing.T) {
	t.Run("return Log value", func(t *testing.T) {
		t.Parallel()

		target := err.New(err.NotFoundError, "user not found", "detail-1", "detail-2")

		got := target.LogValue()
		if got.String() != "[message=user not found: not_found detail=detail-1 | detail-2]" {
			t.Errorf("LogValue() = %v, want %v", got, "[message=user not found: not_found detail=detail-1 | detail-2]")
		}
	})
}

/*
go test -covermode=atomic -count=1 -race -coverprofile=coverage.out $(PACKAGES_TESTS)
	go tool cover -func=coverage.out
*/
