package goapi

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

var errHijackAlreadyWritten = errors.New("goapi: response body already written")

type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.Pusher
	Status() int
	Size() int
}

type responseWriter struct {
	http.ResponseWriter
	status  int
	size    int
	written bool
}

func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.written = true
	}
	n, err := w.ResponseWriter.Write(b)
	if err != nil {
		return n, err
	}
	w.size += n
	return n, err
}
func (w *responseWriter) WriteHeader(statusCode int) {
	if w.written {
		return
	}
	w.ResponseWriter.WriteHeader(statusCode)
	w.status = statusCode
	w.written = true
}

// Hijack implements the http.Hijacker interface.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	// Allow hijacking before any data is written (size == -1) or after headers are written (size == 0),
	// but not after body data is written (size > 0). For compatibility with websocket libraries (e.g., github.com/coder/websocket)
	if w.size > 0 {
		return nil, nil, errHijackAlreadyWritten
	}
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	if w.size < 0 {
		w.size = 0
	}
	return hijacker.Hijack()
}

// Flush implements the http.Flusher interface.
func (w *responseWriter) Flush() {
	if !w.written {
		w.written = true
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *responseWriter) Push(target string, opts *http.PushOptions) error {
	if ps, ok := w.ResponseWriter.(http.Pusher); ok {
		return ps.Push(target, opts)
	}
	return http.ErrNotSupported
}

func (w *responseWriter) Status() int {
	return w.status
}

func (w *responseWriter) Size() int {
	return w.size
}

func (w *responseWriter) reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.size = 0
	w.status = http.StatusOK
	w.written = false
}
