package main

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
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
var embedPosts embed.FS
var posts []fs.DirEntry

//go:embed public/*
var embedPublic embed.FS

type Blog struct {
	Content string
}

func init() {
	var err error
	posts, err = embedPosts.ReadDir("posts")
	if err != nil {
		panic(err)
	}
}

func main() {
	router := chi.NewRouter()
	router.Get("/", getRoot)
	publicHandler := http.FileServerFS(embedPublic)
	router.Get("/public/*", publicHandler.ServeHTTP)
	router.Get("/blog", getBlog)
	router.Get("/blog/{name}", getBlogName)
	router.Get("/api/v0/blog/{name}", apiGetBlogName)
	addr := "localhost:8080"
	slog.Info("serving", "address", fmt.Sprintf("http://%s", addr))
	if err := http.ListenAndServe(addr, router); err != nil {
		panic(err)
	}
}

func unsafe(html string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte(html))
		return err
	})
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	component := templates.Home()
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

func readBlogPost(name string) (string, error) {
	postMd, err := embedPosts.ReadFile(name)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := goldmark.Convert(postMd, &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func InternalServerError(w http.ResponseWriter, _ *http.Request, err error) {
	slog.Error("internal server error", "err", err)
	w.WriteHeader(http.StatusInternalServerError)
}

func getBlog(w http.ResponseWriter, r *http.Request) {
	var cmps []templ.Component
	for _, v := range posts {
		name := v.Name()
		link := templ.URL(path.Join("blog", name))
		cmps = append(cmps, templates.BlogTitle(name, link))
	}
	template := templates.GetBlog(cmps...)
	template.Render(r.Context(), w)
}

func getBlogName(w http.ResponseWriter, r *http.Request) {
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

func apiGetBlogName(w http.ResponseWriter, r *http.Request) {
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
