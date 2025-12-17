package response

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
)

type SSEvent struct {
	w *io.PipeWriter
}

type SSEventData struct {
	Event string
	Data  string
	Id    string
	Retry uint
}

func (s *SSEvent) Write(data SSEventData) {
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
	_, _ = s.w.Write(buf.Bytes())
}

type SSEResponse struct {
	SSEWriter func(s *SSEvent)
}

func (s *SSEResponse) GetStatusCode() int {
	return http.StatusOK
}

func (s *SSEResponse) GetHeader() http.Header {
	return map[string][]string{
		"Content-Type":  {"text/event-stream"},
		"Connection":    {"keep-alive"},
		"Cache-Control": {"no-cache"},
	}
}

func (s *SSEResponse) GetBody() any {
	var r *io.PipeReader
	var w *io.PipeWriter
	if s.SSEWriter != nil {
		r, w = io.Pipe()
		go func() {
			defer func() {
				recover()
			}()
			s.SSEWriter(&SSEvent{w})
			_ = r.Close()
			_ = w.Close()
		}()
	}
	return r
}
