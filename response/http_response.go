package response

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

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
	switch contentType {
	case "application/json":
		buf, err = json.Marshal(h.Body)
	case "application/xml":
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

type exceptInfo struct {
	HttpCode int               `json:"http_code"`
	Header   map[string]string `json:"header"`
	Detail   string            `json:"detail"`
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

func parseHTTPException(errStr string) (httpCode int, header map[string]string, detail string, err error) {
	var res exceptInfo
	if err = json.Unmarshal([]byte(errStr), &res); err != nil {
		return
	}
	httpCode = res.HttpCode
	header = res.Header
	detail = res.Detail
	return
}