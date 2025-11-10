package response

import (
	"encoding/json"
	"net/http"
)

type HTTPResponse[T any] struct {
	HttpCode   int
	HttpHeader http.Header
	HttpBody   T
}

func (h *HTTPResponse[T]) Status() int {
	if h.HttpCode == 0 {
		h.HttpCode = http.StatusOK
	}
	return h.HttpCode
}

func (h *HTTPResponse[T]) Header() http.Header {
	return h.HttpHeader
}

func (h *HTTPResponse[T]) Body() any {
	return h.HttpBody
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
