package response

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// SSE is response
type SSE struct {
	f func(w *SSEWriter)
}

func (*SSE) GetHeader() http.Header {
	return http.Header{
		"Content-Type":  {"text/event-stream"},
		"Connection":    {"keep-alive"},
		"Cache-Control": {"no-cache"},
	}
}

func (s *SSE) GetBody() any {
	var r *io.PipeReader
	var w *io.PipeWriter
	if s.f != nil {
		r, w = io.Pipe()
		go func() {
			defer w.Close()
			defer func() {
				if err := recover(); err != nil {
					wt := &SSEWriter{w: w}
					_ = wt.Write(SSEData{
						Event: "error",
						Data:  fmt.Sprintf("%v", err),
					})
				}
			}()
			s.f(&SSEWriter{w})
		}()
	}
	return r
}

// ReturnSSE return an SSE
func ReturnSSE(f func(w *SSEWriter)) *SSE {
	return &SSE{f: f}
}

type SSEWriter struct {
	w *io.PipeWriter
}

func (s *SSEWriter) Write(data SSEData) (err error) {
	var buf bytes.Buffer
	if data.Event != "" {
		buf.WriteString("event: " + data.Event + "\n")
	}
	buf.WriteString("data: " + data.Data + "\n")
	if data.Id != "" {
		buf.WriteString("id: " + data.Id + "\n")
	}
	if data.Retry > 0 {
		buf.WriteString("retry: " + strconv.Itoa(int(data.Retry)) + "\n")
	}
	buf.WriteString("\n")
	_, err = s.w.Write(buf.Bytes())
	return
}

type SSEData struct {
	Event string
	Data  string
	Id    string
	Retry uint
}
