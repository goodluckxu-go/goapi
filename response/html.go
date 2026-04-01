package response

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
)

// Html is response
type Html struct {
	filename string
	text     string // Highest priority
	data     any
}

// ReturnHtmlByFile return a Html by file
func ReturnHtmlByFile(filename string, data any) *Html {
	return &Html{filename: filename, data: data}
}

// ReturnHtml return a Html
func ReturnHtml(text string, data any) *Html {
	return &Html{text: text, data: data}
}

func (*Html) GetHeader() http.Header {
	return http.Header{
		"Content-Type": []string{"text/html; charset=utf-8"},
	}
}

func (h *Html) GetBody() any {
	var r *io.PipeReader
	var w *io.PipeWriter
	if h.text != "" {
		r, w = io.Pipe()
		go func() {
			defer w.Close()
			defer func() {
				if err := recover(); err != nil {
					_, _ = w.Write([]byte(fmt.Sprintf("error: %v", err)))
				}
			}()
			tmpl := template.Must(template.New("html").Parse(h.text))
			err := tmpl.Execute(w, h.data)
			if err != nil {
				panic(err)
			}
		}()

	} else if h.filename != "" {
		r, w = io.Pipe()
		go func() {
			defer w.Close()
			defer func() {
				if err := recover(); err != nil {
					_, _ = w.Write([]byte(fmt.Sprintf("error: %v", err)))
				}
			}()
			tmpl, err := template.ParseFiles(h.filename)
			if err != nil {
				panic(err)
			}
			err = tmpl.Execute(w, h.data)
			if err != nil {
				panic(err)
			}
		}()
	}
	return r
}
