package main

import (
	"github.com/justinas/alice"
	"github.com/mkailbowdy/ui"
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.FileServerFS(ui.Files))
	mux.HandleFunc("GET /ping", ping)

	dynamic := alice.New(app.sessionManager.LoadAndSave, preventCSRF, app.authenticate)

	protected := dynamic.Append(app.requireAuthentication)

	mux.Handle("GET /{$}", dynamic.ThenFunc(app.home))
	mux.Handle("GET /dashboard", protected.ThenFunc(app.dashboard))
	mux.Handle("GET /url/create", protected.ThenFunc(app.createUrl))
	mux.Handle("POST /url/create", protected.ThenFunc(app.createUrlPost))
	mux.Handle("POST /url/compare", protected.ThenFunc(app.getAndComparePost))
	mux.Handle("POST /url/{id}", protected.ThenFunc(app.showButton))
	mux.Handle("PATCH /url/{id}", protected.ThenFunc(app.updateHashesPost))
	mux.Handle("DELETE /url/{id}", protected.ThenFunc(app.removeSite))

	// User authentication related routes
	mux.Handle("GET /user/signup", dynamic.ThenFunc(app.userSignup))
	mux.Handle("POST /user/signup", dynamic.ThenFunc(app.userSignupPost))
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))
	mux.Handle("POST /user/logout", protected.ThenFunc(app.userLogoutPost))
	// This is a standard chain of middleware used for every request the http server receives.
	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)

	return standard.Then(mux)
}
