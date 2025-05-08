package response

import (
	"encoding/xml"
	json "github.com/json-iterator/go"
	"net/http"
	"strings"
)

type HTTPResponse[T any] struct {
	HttpCode int
	Header   map[string]string
	Cookie   map[string]string
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
	isSet := false
	for k, _ := range h.Header {
		if strings.ToLower(k) == "content-type" {
			isSet = true
			break
		}
	}
	if !isSet {
		h.Header["Content-Type"] = contentType
	}
}

func (h *HTTPResponse[T]) Write(w http.ResponseWriter) {
	for k, v := range h.Header {
		w.Header().Set(k, v)
	}
	for k, v := range h.Cookie {
		w.Header().Add("Set-Cookie", k+"="+v)
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
	switch contentType {
	case "application/json":
		buf, err = json.Marshal(h.Body)
	case "application/xml":
		buf, err = xml.Marshal(h.Body)
		buf = append([]byte(`<?xml version="1.0" encoding="UTF-8"?>`), buf...)
	default:
		var anyVal any = h.Body
		switch val := anyVal.(type) {
		case []byte:
			buf = val
		case string:
			buf = []byte(val)
		}
	}
	if err != nil {
		HTTPException(http.StatusInternalServerError, err.Error())
	}
	_, _ = w.Write(buf)
}

type exceptInfo struct {
	HttpCode int    `json:"http_code"`
	Detail   string `json:"detail"`
}

func HTTPException(httpCode int, detail string) {
	res := exceptInfo{
		HttpCode: httpCode,
		Detail:   detail,
	}
	buf, _ := json.Marshal(&res)
	panic(string(buf))
}
