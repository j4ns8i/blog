package main

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"path"
	"text/template"

	"github.com/a-h/templ"
	"github.com/j4ns8i/blog/templates"
	"github.com/yuin/goldmark"
)

//go:fmt echo hello world

//go:generate go run github.com/a-h/templ/cmd/templ generate

//go:embed posts/*
var blogs embed.FS

//go:embed main.tmpl
var templateSource string

type Blog struct {
	Content string
}

var tmpl *template.Template

func init() {
	var err error
	tmpl, err = template.New("main").Parse(templateSource)
	if err != nil {
		panic(err)
	}
}

func main() {
	http.HandleFunc("/blog/{name}", getBlog)
	http.ListenAndServe(":8080", nil)
}

func unsafe(html string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte(html))
		return err
	})
}

func getBlog(w http.ResponseWriter, r *http.Request) {
	name := path.Join("posts", r.PathValue("name"))
	postMd, err := blogs.ReadFile(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	var buf bytes.Buffer
	if err := goldmark.Convert(postMd, &buf); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	component := templates.Blog(unsafe(buf.String()))
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}
