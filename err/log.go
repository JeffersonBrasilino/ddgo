// Intent: integrate package errors with the standard structured logger, log/slog.
//
// Objective: convert any error, whether a *DdgoError chain or a plain error, into a single
// slog attribute whose message and details are flattened to one line, keeping log
// aggregation tools from splitting an entry on embedded line breaks.

package err

import (
	"log/slog"
	"strings"
)

const (
	separator = " | "
	// lineBreaks lists every character treated as a line break by flatten: LF, CR, NEL and
	// the Unicode line/paragraph separators, all of which can split a log entry.
	lineBreaks = "\n\r\u0085\u2028\u2029"
)

// ToSlogAttr converts e into a slog attribute keyed "error", ready to be passed to any
// slog logging call.
//
// Parameters:
//   - e: the error to convert; it may be a *DdgoError chain, any other error, or nil.
//
// Returns a group attribute containing:
//   - "message": the error message, with line breaks (LF, CR, CRLF and the Unicode line
//     separators) replaced by " | ";
//   - "detail": only when the error tree contains details, those of every *DdgoError
//     found (including behind fmt.Errorf %w and errors.Join), joined by " | ", with line
//     breaks replaced the same way.
//
// Details are logged verbatim apart from line breaks; see [New] about keeping personal
// data out of them.
//
// When e is nil, it returns the zero [slog.Attr], which slog handlers ignore.
func ToSlogAttr(e error) slog.Attr {
	if e == nil {
		return slog.Attr{}
	}

	attrs := []slog.Attr{
		slog.String("message", flatten(e.Error())),
	}

	if details := collectDetails(e, nil); len(details) > 0 {
		attrs = append(attrs, slog.String("detail", flatten(strings.Join(details, separator))))
	}

	return slog.Attr{Key: "error", Value: slog.GroupValue(attrs...)}
}

// flatten collapses text into a single line, replacing every run of line breaks (see
// lineBreaks) with separator.
func flatten(text string) string {
	if !strings.ContainsAny(text, lineBreaks) {
		return text
	}

	return strings.Join(strings.FieldsFunc(text, func(char rune) bool {
		return strings.ContainsRune(lineBreaks, char)
	}), separator)
}
