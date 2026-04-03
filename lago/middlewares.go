package lago

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"gorm.io/gorm"
)

type DBMiddleware struct {
	DB *gorm.DB
}

func (m DBMiddleware) Next(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "$db", m.DB)))
	})
}

type LoggingMiddleware struct{}

func (LoggingMiddleware) Next(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		slog.Info("http_request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", wrapped.status),
			slog.Duration("latency", time.Since(start)),
			slog.String("ip", r.RemoteAddr),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

type EnvironmentMiddleware struct{}

func (EnvironmentMiddleware) Next(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		envMap := map[string]string{}
		cookie, err := r.Cookie("environment")
		if err != nil {
			slog.Error("MiddlewareEnvironment: failed to get environment cookie", "error", err)
			setEmptyEnvironmentCookie(w, "environment")
		} else {
			decoded, err := url.QueryUnescape(cookie.Value)
			if err == nil {
				if err := json.Unmarshal([]byte(decoded), &envMap); err != nil {
					slog.Error("MiddlewareEnvironment: failed to unmarshal environment cookie", "error", err, "cookie", cookie.Value)
					setEmptyEnvironmentCookie(w, "environment")
				}
			} else {
				slog.Error("Error while decoding cookie value", "error", err)
				setEmptyEnvironmentCookie(w, "environment")
			}
		}
		ctx := context.WithValue(r.Context(), "$environment", envMap)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func setEmptyEnvironmentCookie(w http.ResponseWriter, cookieName string) {
	http.SetCookie(w, &http.Cookie{
		Name:  cookieName,
		Value: url.QueryEscape("{}"),
		Path:  "/",
	})
}

type HtmxBoostMiddleware struct{}

func (HtmxBoostMiddleware) Next(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isBoosted := r.Header.Get("HX-Boosted") == "true"
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "isHtmxBoosted", isBoosted)))
	})
}

type CacheDisableMiddleware struct{}

func (CacheDisableMiddleware) Next(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		w.Header().Del("ETag")
		w.Header().Del("Last-Modified")
		next.ServeHTTP(w, r)
	})
}

type ResponseLoggerMiddleware struct{}

func (ResponseLoggerMiddleware) Next(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rww := NewResponseWriterWrapper(w)
		w.Header()
		defer func() {
			slog.Info(
				fmt.Sprintf(
					"[Execution time: %v] [Response: %s]",
					time.Since(start),
					rww.String(),
				))
		}()
		next.ServeHTTP(rww, r)
	})
}

// ResponseWriterWrapper struct is used to log the response
type ResponseWriterWrapper struct {
	w          *http.ResponseWriter
	body       *bytes.Buffer
	statusCode *int
}

// NewResponseWriterWrapper static function creates a wrapper for the http.ResponseWriter
func NewResponseWriterWrapper(w http.ResponseWriter) ResponseWriterWrapper {
	var buf bytes.Buffer
	var statusCode int = 200
	return ResponseWriterWrapper{
		w:          &w,
		body:       &buf,
		statusCode: &statusCode,
	}
}

func (rww ResponseWriterWrapper) Write(buf []byte) (int, error) {
	rww.body.Write(buf)
	return (*rww.w).Write(buf)
}

// Header function overwrites the http.ResponseWriter Header() function
func (rww ResponseWriterWrapper) Header() http.Header {
	return (*rww.w).Header()
}

// WriteHeader function overwrites the http.ResponseWriter WriteHeader() function
func (rww ResponseWriterWrapper) WriteHeader(statusCode int) {
	(*rww.statusCode) = statusCode
	(*rww.w).WriteHeader(statusCode)
}

func (rww ResponseWriterWrapper) String() string {
	var buf bytes.Buffer

	buf.WriteString("Response:")

	buf.WriteString("Headers:")
	for k, v := range (*rww.w).Header() {
		buf.WriteString(fmt.Sprintf("%s: %v", k, v))
	}

	buf.WriteString(fmt.Sprintf(" Status Code: %d", *(rww.statusCode)))

	buf.WriteString("Body")
	buf.WriteString(rww.body.String())
	return buf.String()
}
