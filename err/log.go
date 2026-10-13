package err

import (
	"errors"
	"log/slog"
	"strings"
)

const separator = " | "

func ToSlogAttr(e error) slog.Attr {
	if e == nil {
		return slog.Attr{}
	}

	attrs := []slog.Attr{
		slog.String("message", flatten(makeErrorForLogs(e, 0))),
	}

	if ddgoErr, ok := errors.AsType[*DdgoErr](e); ok {
		attrs = append(
			attrs,
			slog.String("detail", strings.Join(ddgoErr.mapErrorDetails(0), separator)),
		)
	}

	return slog.Attr{Key: "error", Value: slog.GroupValue(attrs...)}
}

func flatten(text string) string {
	if !strings.Contains(text, "\n") {
		return text
	}

	return strings.Join(strings.FieldsFunc(text, func(char rune) bool {
		return char == '\n'
	}), separator)
}

func makeErrorForLogs(e error, depth int) string {

	if ddgoErr, ok := errors.AsType[*DdgoErr](e); ok {
		return ddgoErr.mapErrorCauses(depth)
	}

	return e.Error()
}
