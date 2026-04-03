package sqlagent

import "log/slog"

// logError logs err with slog.Error when err != nil. attrs are alternating key/value pairs.
func logError(msg string, err error, attrs ...any) {
	if err == nil {
		return
	}
	args := make([]any, 0, 2+len(attrs))
	args = append(args, "error", err)
	args = append(args, attrs...)
	slog.Error(msg, args...)
}
