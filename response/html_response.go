package response

import (
	"html/template"
	"net/http"
)

type HTMLResponse struct {
	Filename string
	Html     string // Highest priority
	Data     any
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
	var tmpl *template.Template
	var err error
	if len(h.Html) > 0 {
		tmpl = template.Must(template.New("html").Parse(h.Html))
	} else {
		tmpl, err = template.ParseFiles(h.Filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	err = tmpl.Execute(w, h.Data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
