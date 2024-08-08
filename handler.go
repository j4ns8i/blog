package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"sync"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/j4ns8i/blog/templates"
	"github.com/yuin/goldmark"
)

func Unsafe(html string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte(html))
		return err
	})
}

func InternalServerError(w http.ResponseWriter, _ *http.Request, err error) {
	slog.Error("internal server error", "err", err)
	w.WriteHeader(http.StatusInternalServerError)
}

type handler struct {
	chi.Router

	renderedPosts sync.Map
}

func newHandler() *handler {
	var h = new(handler)
	publicHandler := http.FileServerFS(embedPublic)
	h.Router = chi.NewRouter()
	h.Router.Get("/", getRoot)
	h.Router.Get("/public/*", publicHandler.ServeHTTP)
	h.Router.Get("/blog", h.GetBlog)
	h.Router.Get("/blog/{name}", h.GetBlogName)
	return h
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	component := templates.Home()
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

// blog renders the blog post, caching rendered outputs
// TODO: cache the components instead
func (h *handler) blog(name string) (string, error) {
	if html, ok := h.renderedPosts.Load(name); ok {
		return html.(string), nil
	}
	html, err := h.renderBlog(name)
	if err != nil {
		return "", err
	}
	h.renderedPosts.Store(name, html)
	return html, nil
}

func (h *handler) renderBlog(name string) (string, error) {
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

func (h *handler) GetBlog(w http.ResponseWriter, r *http.Request) {
	var cmps []templ.Component
	// TODO: cache
	for _, v := range posts {
		name := v.Name()
		link := templ.URL(path.Join("blog", name))
		cmps = append(cmps, templates.BlogTitle(name, link))
	}
	template := templates.GetBlog(cmps...)
	template.Render(r.Context(), w)
}

func (h *handler) GetBlogName(w http.ResponseWriter, r *http.Request) {
	name := path.Join("posts", r.PathValue("name"))
	blogHtml, err := h.blog(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// TODO: Return a proper 404 page
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	blog := templates.Blog(Unsafe(blogHtml))
	w.Header().Set("Content-Type", "text/html")
	blog.Render(r.Context(), w)
}

func (h *handler) apiGetBlogName(w http.ResponseWriter, r *http.Request) {
	name := path.Join("posts", r.PathValue("name"))
	blogHtml, err := h.blog(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	blog := templates.ApiBlog(Unsafe(blogHtml))
	w.Header().Set("Content-Type", "text/html")
	blog.Render(r.Context(), w)
}
