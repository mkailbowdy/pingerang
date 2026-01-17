package main

import (
	"net/http"
)

func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))
	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /contact/1", app.contact)
	mux.HandleFunc("GET /contact/1/view", app.viewForm)
	mux.HandleFunc("GET /contact/1/edit", app.editForm)
	mux.HandleFunc("POST /url", app.urlCreatePost)
	mux.HandleFunc("POST /url/compare", app.urlComparePost)

	return mux
}
