package err_test

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/jeffersonbrasilino/ddgo/err"
)

func TestLog_ToSlogAttr(t *testing.T) {
	t.Run("returns slog attrs", func(t *testing.T) {
		t.Parallel()
		er := err.New(err.NotFoundError, "user not found", "detail-1", "detail-2")
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
}
