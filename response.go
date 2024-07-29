package goapi

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
)

type Response interface {
	Bytes() []byte
	GetBody() any
	GetHttpCode() int
	GetHeaders() map[string]string
}

type HTTPResponse[T any] struct {
	HttpCode int
	Header   map[string]string
	Body     T
}

func (h *HTTPResponse[T]) Bytes() []byte {
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
	case "application/octet-stream":
		var anyVal any = h.Body
		if val, ok := anyVal.([]byte); ok {
			buf = val
		}
	}
	if err != nil {
		HTTPException(http.StatusInternalServerError, err.Error())
	}
	return buf
}

func (h *HTTPResponse[T]) GetBody() any {
	return h.Body
}

func (h *HTTPResponse[T]) GetHttpCode() int {
	return h.HttpCode
}

func (h *HTTPResponse[T]) GetHeaders() map[string]string {
	return h.Header
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

func (h *FileResponse) Bytes() []byte {
	return h.Body
}

func (h *FileResponse) GetBody() any {
	return h.Body
}

func (h *FileResponse) GetHttpCode() int {
	return 200
}

func (h *FileResponse) GetHeaders() map[string]string {
	return map[string]string{
		"Content-Type":        "application/octet-stream",
		"Content-Disposition": fmt.Sprintf("attachment; filename=\"%v\"", h.Filename),
	}
}
