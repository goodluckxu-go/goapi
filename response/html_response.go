package response

import (
	"html/template"
	"net/http"
)

type HTMLResponse struct {
	Filename string
	Data     any
	Html     []byte // Highest priority
}

func (h *HTMLResponse) GetBody() any {
	return ""
}

func (h *HTMLResponse) GetContentType() string {
	return "text/html"
}

func (h *HTMLResponse) SetContentType(contentType string) {
}

func (h *HTMLResponse) Write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", h.GetContentType())
	if h.Html != nil {
		_, _ = w.Write(h.Html)
		return
	}
	tmpl, err := template.ParseFiles(h.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, h.Data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
