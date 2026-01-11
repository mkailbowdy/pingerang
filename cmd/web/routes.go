package main

import (
	"net/http"
)

func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("POST /url", app.urlCreatePost)
	mux.HandleFunc("POST /url/compare", app.urlComparePost)

	return mux
}
