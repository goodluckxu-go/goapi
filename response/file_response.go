package response

import (
	"fmt"
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

func (h *FileResponse[T]) GetBody() any {
	return h.Body
}

func (h *FileResponse[T]) GetContentType() string {
	return h.t.ContentType()
}

func (h *FileResponse[T]) SetContentType(contentType string) {
}

func (h *FileResponse[T]) Write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", h.GetContentType())
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", h.Filename))
	w.WriteHeader(200)
	_, _ = w.Write(h.Body)
}
