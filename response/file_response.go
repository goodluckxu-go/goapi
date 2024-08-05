package response

import (
	"fmt"
	"net/http"
)

type FileResponse struct {
	Filename string
	Body     []byte
}

func (h *FileResponse) GetBody() any {
	return h.Body
}

func (h *FileResponse) GetContentType() string {
	return "application/octet-stream"
}

func (h *FileResponse) SetContentType(contentType string) {
}

func (h *FileResponse) Write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", h.GetContentType())
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", h.Filename))
	w.WriteHeader(200)
	_, _ = w.Write(h.Body)
}
