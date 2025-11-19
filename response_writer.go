package goapi

import (
	"bufio"
	"net"
	"net/http"
)

type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	Status() int
	Size() int
}

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	if err != nil {
		return n, err
	}
	w.size += n
	return n, err
}
func (w *responseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.status = statusCode
}

// Hijack implements the http.Hijacker interface.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

// Flush implements the http.Flusher interface.
func (w *responseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *responseWriter) Status() int {
	return w.status
}

func (w *responseWriter) Size() int {
	return w.size
}

func (w *responseWriter) reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.status = http.StatusOK
}
