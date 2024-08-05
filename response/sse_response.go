package response

import (
	"bytes"
	"net/http"
	"strconv"
)

type SSEvent struct {
	w http.ResponseWriter
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
		buf.WriteString("event: " + data.Event + "\n\n")
	}
	buf.WriteString("data: " + data.Data + "\n\n")
	if data.Id != "" {
		buf.WriteString("id: " + data.Id + "\n\n")
	}
	if data.Retry > 0 {
		buf.WriteString("retry: " + strconv.Itoa(int(data.Retry)) + "\n\n")
	}
	_, _ = s.w.Write(buf.Bytes())
	if f, ok := s.w.(http.Flusher); ok {
		f.Flush()
	}
}

type SSEResponse struct {
	SSEWriter func(s *SSEvent)
}

func (s *SSEResponse) GetBody() any {
	return ""
}

func (s *SSEResponse) GetContentType() string {
	return "text/event-stream"
}

func (s *SSEResponse) SetContentType(contentType string) {
}

func (s *SSEResponse) Write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", s.GetContentType())
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(200)
	s.SSEWriter(&SSEvent{w: w})
}
