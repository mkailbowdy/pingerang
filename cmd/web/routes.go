package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	dynamic := alice.New(app.sessionManager.LoadAndSave)

	mux.Handle("GET /{$}", dynamic.ThenFunc(app.home))
	mux.Handle("GET /dashboard", dynamic.ThenFunc(app.dashboard))
	mux.Handle("GET /url/create", dynamic.ThenFunc(app.createUrl))
	mux.Handle("POST /url/create", dynamic.ThenFunc(app.createUrlPost))
	mux.Handle("POST /url/compare", dynamic.ThenFunc(app.getAndComparePost))
	mux.Handle("POST /url/{id}", dynamic.ThenFunc(app.showButton))
	mux.Handle("PATCH /url/{id}", dynamic.ThenFunc(app.updateHashesPost))

	// This is a standard chain of middleware used for every request the http server receives.
	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)

	return standard.Then(mux)
}
