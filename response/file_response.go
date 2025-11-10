package response

import (
	"net/http"
)

type FileInterface interface {
	ContentType() string
}

type FileResponse[T FileInterface] struct {
	Filename string
	HttpBody []byte
	t        T
}

func (f *FileResponse[T]) Status() int {
	return http.StatusOK
}

func (f *FileResponse[T]) Header() http.Header {
	return map[string][]string{
		"Content-Type":        {f.t.ContentType()},
		"Content-Disposition": {"attachment; filename=\"" + f.Filename + "\""},
	}
}

func (f *FileResponse[T]) Body() any {
	return f.HttpBody
}
