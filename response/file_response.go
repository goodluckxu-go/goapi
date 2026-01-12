package response

import (
	"net/http"
	"net/url"
)

type FileInterface interface {
	ContentType() string
}

type FileResponse[T FileInterface] struct {
	Filename string
	Body     []byte
	t        T
}

func (f *FileResponse[T]) GetStatusCode() int {
	return http.StatusOK
}

func (f *FileResponse[T]) GetHeader() http.Header {
	return map[string][]string{
		"Content-Type":        {f.t.ContentType()},
		"Content-Disposition": {"attachment; filename=\"" + url.PathEscape(f.Filename) + "\""},
	}
}

func (f *FileResponse[T]) GetBody() any {
	return f.Body
}
