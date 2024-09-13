package response

import (
	"net/http"
)

type TextResponse struct {
	Header map[string]string
	Cookie map[string]string
	Body   []byte
}

func (h *TextResponse) GetBody() any {
	return h.Body
}

func (h *TextResponse) GetContentType() string {
	return "text/plain"
}

func (h *TextResponse) SetContentType(contentType string) {
}

func (h *TextResponse) Write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", h.GetContentType())
	for k, v := range h.Header {
		w.Header().Set(k, v)
	}
	for k, v := range h.Cookie {
		w.Header().Add("Set-Cookie", k+"="+v)
	}
	w.WriteHeader(200)
	_, _ = w.Write(h.Body)
}
