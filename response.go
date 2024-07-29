package goapi

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
)

type Response interface {
	GetBody() any
	GetContentType() string
	SetContentType(contentType string)
	Write(w http.ResponseWriter)
}

type HTTPResponse[T any] struct {
	HttpCode int
	Header   map[string]string
	Body     T
}

func (h *HTTPResponse[T]) GetBody() any {
	return h.Body
}

func (h *HTTPResponse[T]) GetContentType() string {
	if h.Header == nil {
		return ""
	}
	return h.Header["Content-Type"]
}

func (h *HTTPResponse[T]) SetContentType(contentType string) {
	if h.Header == nil {
		h.Header = map[string]string{}
	}
	h.Header["Content-Type"] = contentType
}

func (h *HTTPResponse[T]) Write(w http.ResponseWriter) {
	for k, v := range h.Header {
		w.Header().Set(k, v)
	}
	if h.HttpCode == 0 {
		h.HttpCode = 200
	}
	w.WriteHeader(h.HttpCode)
	// get body bytes
	contentType := ""
	if h.Header != nil {
		contentType = h.Header["Content-Type"]
	}
	var buf []byte
	var err error
	switch MediaType(contentType) {
	case JSON:
		buf, err = json.Marshal(h.Body)
	case XML:
		buf, err = xml.Marshal(h.Body)
	default:
		var anyVal any = h.Body
		if val, ok := anyVal.([]byte); ok {
			buf = val
		}
	}
	if err != nil {
		HTTPException(http.StatusInternalServerError, err.Error())
	}
	_, _ = w.Write(buf)
}

func HTTPException(httpCode int, detail string, headers ...map[string]string) {
	header := map[string]string{}
	for _, item := range headers {
		for k, v := range item {
			header[k] = v
		}
	}
	res := exceptInfo{
		HttpCode: httpCode,
		Header:   header,
		Detail:   detail,
	}
	buf, _ := json.Marshal(&res)
	panic(string(buf))
}

type FileResponse struct {
	Filename string
	Body     []byte
}

func (h *FileResponse) GetBody() any {
	return h.Body
}

func (h *FileResponse) GetContentType() string {
	return "application/octet-stream"
}

func (h *FileResponse) SetContentType(contentType string) {
}

func (h *FileResponse) Write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", h.GetContentType())
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", h.Filename))
	w.WriteHeader(200)
	_, _ = w.Write(h.Body)
}

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
