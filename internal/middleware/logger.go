package middleware

import (
	"log/slog"
	"net/http"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Request log:",
			slog.String("host", r.Host),
			slog.String("uri", r.RequestURI),
			slog.String("protocol", r.Proto),
			slog.String("method", r.Method),
			slog.String("remote_address", r.RemoteAddr),
			slog.String("real_IP", r.Header.Get("X-Real-IP")),
			slog.String("request_id", GetReqID(r.Context())),
		)

		next.ServeHTTP(w, r)

		slog.Info("Response log:",
			slog.String("status", w.Header().Get("Status Code")),
			slog.String("content-type", w.Header().Get("Content-Type")),
		)
	})
}
