// Package log provides a context-aware structured logger.
package log

import (
	"context"
	"fmt"
	"log/slog"
)

// Logger is a structured log handler.
type Logger []slog.Attr

// std is the default Logger.
var std Logger

// With returns a Logger that includes the given attributes in each log entry.
func With(args ...any) Logger {
	n := len(args)
	if n == 0 {
		return Logger{}
	}
	a := make(Logger, 0, n)
	var r slog.Record
	r.Add(args...)
	r.Attrs(func(attr slog.Attr) bool {
		a = append(a, attr)
		return true
	})
	return a
}

// Debugf outputs a debug level log entry with the attributes set on a.
func (l Logger) Debugf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, slog.LevelDebug, format, args...)
}

// Infof outputs an info level log entry with the attributes set on a.
func (l Logger) Infof(ctx context.Context, format string, args ...any) {
	l.logf(ctx, slog.LevelInfo, format, args...)
}

// Warnf outputs a warning level log entry with the attributes set on a.
func (l Logger) Warnf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, slog.LevelWarn, format, args...)
}

// Errorf ouputs an error level log entry with the attributes set on a.
func (l Logger) Errorf(ctx context.Context, format string, args ...any) {
	l.logf(ctx, slog.LevelError, format, args...)
}

// Debugf outputs a debug level log entry.
func Debugf(ctx context.Context, format string, args ...any) {
	std.logf(ctx, slog.LevelDebug, format, args...)
}

// Infof outputs an info level log entry.
func Infof(ctx context.Context, format string, args ...any) {
	std.logf(ctx, slog.LevelInfo, format, args...)
}

// Warnf outputs a warning level log entry.
func Warnf(ctx context.Context, format string, args ...any) {
	std.logf(ctx, slog.LevelWarn, format, args...)
}

// Errorf ouputs an error level log entry.
func Errorf(ctx context.Context, format string, args ...any) {
	std.logf(ctx, slog.LevelError, format, args...)
}

func (l Logger) logf(ctx context.Context, level slog.Level, format string, args ...any) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	FromContext(ctx).LogAttrs(ctx, level, msg, l...)
}

type ctxKey struct{}

// FromContext returns the [slog.Logger] from the given context if it exists,
// otherwise it returns [slog.Default].
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// SetContext returns a derived context that contains the given logger.
func SetContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}
