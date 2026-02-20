package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jessesomerville/yodahunters/internal/log"
)

func Logger(ctx context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.With(
			reqToLogAttr(r),
			slog.String("latency", time.Since(start).String()),
		).Infof(ctx, "received request")
	})
}

func reqToLogAttr(r *http.Request) slog.Attr {
	return slog.GroupAttrs("req",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.Int64("content-length", r.ContentLength),
		slog.String("ip", r.RemoteAddr),
	)
}
