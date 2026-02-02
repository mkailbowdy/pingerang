package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))

	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))
	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /dashboard", app.dashboard)
	mux.HandleFunc("GET /url/create", app.createUrl)
	mux.HandleFunc("POST /url/create", app.createUrlPost)
	mux.HandleFunc("POST /url/compare", app.getAndComparePost)
	mux.HandleFunc("POST /url/{id}", app.showButton)
	mux.HandleFunc("PATCH /url/{id}", app.updateHashesPost)

	// This is a standard chain of middleware used for every request the http server receives.
	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)

	return standard.Then(mux)
}