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

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/j4ns8i/blog/templates"
	"github.com/yuin/goldmark"
)

//go:fmt echo hello world

//go:generate go run github.com/a-h/templ/cmd/templ generate

//go:embed posts/*
var posts embed.FS

//go:embed public/styles.css
var stylesheet []byte

type Blog struct {
	Content string
}

func main() {
	router := chi.NewRouter()
	router.Get("/", getRoot)
	router.Get("/public/styles.css", getStylesheet)
	router.Get("/blog/{name}", getBlog)
	router.Get("/api/v0/blog/{name}", apiGetBlog)
	http.ListenAndServe(":8080", router)
}

func unsafe(html string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte(html))
		return err
	})
}

func getStylesheet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	w.Write(stylesheet)
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	component := templates.Root()
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

func readBlogPost(name string) (string, error) {
	postMd, err := posts.ReadFile(name)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := goldmark.Convert(postMd, &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func getBlog(w http.ResponseWriter, r *http.Request) {
	name := path.Join("posts", r.PathValue("name"))
	blogHtml, err := readBlogPost(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// TODO: Return a proper 404 page
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	blog := templates.Blog(unsafe(blogHtml))
	w.Header().Set("Content-Type", "text/html")
	blog.Render(r.Context(), w)
}

func apiGetBlog(w http.ResponseWriter, r *http.Request) {
	name := path.Join("posts", r.PathValue("name"))
	blogHtml, err := readBlogPost(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	blog := templates.ApiBlog(unsafe(blogHtml))
	w.Header().Set("Content-Type", "text/html")
	blog.Render(r.Context(), w)
}
