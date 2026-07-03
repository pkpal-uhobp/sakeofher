package middleware

import "net/http"

type StatusResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func NewStatusResponseWriter(w http.ResponseWriter) *StatusResponseWriter {
	return &StatusResponseWriter{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}
}

func (w *StatusResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
