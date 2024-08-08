package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
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
	h := newHandler()
	addr := "localhost:8080"
	slog.Info("serving", "address", fmt.Sprintf("http://%s", addr))
	if err := http.ListenAndServe(addr, h); err != nil {
		panic(err)
	}
}
