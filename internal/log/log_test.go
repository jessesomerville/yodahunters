package log

import (
	"bytes"
	"context"
	"log"
	"log/slog"
	"testing"
)

func testSlogHandler(buf *bytes.Buffer, lvl slog.Level) *slog.Logger {
	h := slog.NewTextHandler(buf, &slog.HandlerOptions{
		Level: lvl,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			// Remove time attribute to prevent flaky tests.
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
	return slog.New(h)
}

func TestFromContext(t *testing.T) {
	var buf bytes.Buffer
	l := testSlogHandler(&buf, slog.LevelDebug).With("TestFromContext", "test_attribute")
	ctx := SetContext(context.Background(), l)

	Debugf(ctx, "msg")
	want := "level=DEBUG msg=msg TestFromContext=test_attribute\n"
	if got := buf.String(); got != want {
		t.Errorf("Debugf(ctx, \"msg\")\n  got=%q\n  want=%q", got, want)
	}
	buf.Reset()

	Infof(ctx, "msg")
	want = "level=INFO msg=msg TestFromContext=test_attribute\n"
	if got := buf.String(); got != want {
		t.Errorf("Infof(ctx, \"msg\")\n  got=%q\n  want=%q", got, want)
	}
	buf.Reset()

	Infof(ctx, "%s %d", "msg", 123)
	want = "level=INFO msg=\"msg 123\" TestFromContext=test_attribute\n"
	if got := buf.String(); got != want {
		t.Errorf("Infof(ctx, \"%%s %%d\", \"msg\", 123)\n  got=%q\n  want=%q", got, want)
	}
	buf.Reset()

	Warnf(ctx, "msg")
	want = "level=WARN msg=msg TestFromContext=test_attribute\n"
	if got := buf.String(); got != want {
		t.Errorf("Warnf(ctx, \"msg\")\n  got=%q\n  want=%q", got, want)
	}
	buf.Reset()

	Errorf(ctx, "msg")
	want = "level=ERROR msg=msg TestFromContext=test_attribute\n"
	if got := buf.String(); got != want {
		t.Errorf("Errorf(ctx, \"msg\")\n  got=%q\n  want=%q", got, want)
	}
	buf.Reset()

	// Disable debug logging
	l = testSlogHandler(&buf, slog.LevelInfo)
	ctx = SetContext(context.Background(), l)
	Debugf(ctx, "msg")
	want = ""
	if got := buf.String(); got != want {
		t.Errorf("Debugf(ctx, \"msg\")\n  got=%q\n  want=%q", got, want)
	}
}

func TestWith(t *testing.T) {
	currentLogger := slog.Default()
	currentLogWriter := log.Writer()
	currentLogFlags := log.Flags()
	t.Cleanup(func() {
		slog.SetDefault(currentLogger)
		log.SetOutput(currentLogWriter)
		log.SetFlags(currentLogFlags)
	})

	var buf bytes.Buffer
	l := testSlogHandler(&buf, slog.LevelDebug)
	slog.SetDefault(l)

	ctx := context.Background()

	With().Infof(ctx, "msg")
	want := "level=INFO msg=msg\n"
	if got := buf.String(); got != want {
		t.Errorf("With().Infof(ctx, \"msg\")\n  got=%q\n  want=%q", got, want)
	}
	buf.Reset()

	w := With("foo", 123)
	w.Debugf(ctx, "msg")
	want = "level=DEBUG msg=msg foo=123\n"
	if got := buf.String(); got != want {
		t.Errorf("With(\"foo\", 123).Debugf(ctx, \"msg\")\n  got=%q\n  want=%q", got, want)
	}
	buf.Reset()

	w.Infof(ctx, "msg")
	want = "level=INFO msg=msg foo=123\n"
	if got := buf.String(); got != want {
		t.Errorf("With(\"foo\", 123).Infof(ctx, \"msg\")\n  got=%q\n  want=%q", got, want)
	}
	buf.Reset()

	w.Warnf(ctx, "msg")
	want = "level=WARN msg=msg foo=123\n"
	if got := buf.String(); got != want {
		t.Errorf("With(\"foo\", 123).Warnf(ctx, \"msg\")\n  got=%q\n  want=%q", got, want)
	}
	buf.Reset()

	w.Errorf(ctx, "msg")
	want = "level=ERROR msg=msg foo=123\n"
	if got := buf.String(); got != want {
		t.Errorf("With(\"foo\", 123).Errorf(ctx, \"msg\")\n  got=%q\n  want=%q", got, want)
	}
}
