package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/labstack/gommon/color"
)

type responseRecoder struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseRecoder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func loggingLine(method, path string, statusCode int, duration time.Duration) {
	level := "INFO"
	colorFn := color.Green

	switch {
	case statusCode >= 500:
		level = "ERROR"
		colorFn = color.Red
	case statusCode >= 400:
		level = "WARN"
		colorFn = color.Yellow
	case statusCode >= 300:
		level = "INFO"
		colorFn = color.Cyan
	default:
		colorFn = color.Green
	}

	log.Printf("%s %s %s %d %s",
		colorFn(level),
		method,
		path,
		statusCode,
		duration,
	)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		rec := &responseRecoder{w, http.StatusOK}

		next.ServeHTTP(rec, r)
		elapsedTime := time.Since(startTime)

		loggingLine(r.Method, r.URL.Path, rec.statusCode, elapsedTime)
	})
}
