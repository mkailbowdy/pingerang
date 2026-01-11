package main

import (
	"log"
	"net/http"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from Pingerang"))
}

func (app *application) urlCreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err.Error())
	}
	url := r.PostForm.Get("url")
	urlhash, pagehash := drive(url)
	_, err = app.sites.Insert(url, urlhash, pagehash)
	if err != nil {
		log.Fatal(err.Error())
	}
}
