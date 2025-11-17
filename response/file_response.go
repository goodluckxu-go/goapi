package response

import (
	"net/http"
)

type FileInterface interface {
	ContentType() string
}

type FileResponse[T FileInterface] struct {
	Filename string
	Body     []byte
	t        T
}

func (f *FileResponse[T]) HttpStatus() int {
	return http.StatusOK
}

func (f *FileResponse[T]) HttpHeader() http.Header {
	return map[string][]string{
		"Content-Type":        {f.t.ContentType()},
		"Content-Disposition": {"attachment; filename=\"" + f.Filename + "\""},
	}
}

func (f *FileResponse[T]) HttpBody() any {
	return f.Body
}
