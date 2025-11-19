package response

import (
	"encoding/json"
	"net/http"
)

type HTTPResponse[T any] struct {
	HttpCode int
	Header   http.Header
	Body     T
}

func (h *HTTPResponse[T]) GetStatusCode() int {
	if h.HttpCode == 0 {
		h.HttpCode = http.StatusOK
	}
	return h.HttpCode
}

func (h *HTTPResponse[T]) GetHeader() http.Header {
	return h.Header
}

func (h *HTTPResponse[T]) GetBody() any {
	return h.Body
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
