package goapi

import (
	"encoding/json"
	"net/http"
)

type Response interface {
	ResponseStatusCode
	ResponseHeader
	ResponseBody
}

type ResponseStatusCode interface {
	GetStatusCode() int
}

type ResponseHeader interface {
	GetHeader() http.Header
}

type ResponseBody interface {
	GetBody() any
}

func HTTPException(httpCode int, detail string) {
	res := exceptJson{
		HttpCode: httpCode,
		Detail:   detail,
	}
	buf, _ := json.Marshal(&res)
	panic(string(buf))
}
