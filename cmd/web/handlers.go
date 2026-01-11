package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/mkailbowdy/internal/models"
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
	urlhash, pagehash := driveHash(url)
	_, err = app.sites.Insert(url, urlhash, pagehash)
	if err != nil {
		fmt.Printf("%s", err.Error())
		return
	}
}

func (app *application) urlComparePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err.Error())
	}
	url := r.PostForm.Get("url")
	urlhash, pagehash := driveHash(url)
	storedHash, err := app.sites.Get(urlhash)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
			fmt.Printf("This url has not been registered.\n")
			return
		} else {
			fmt.Printf("Error: %s", err.Error())
		}
	}

	if storedHash == pagehash {
		fmt.Printf("No changes on the page.\n")
		return
	}

	fmt.Printf("The page has changed: %s\n", url)
	err = app.sites.Update(urlhash, pagehash)
	if err != nil {
		fmt.Printf("%s", err.Error())
		return
	}
	fmt.Printf("Record updated.\n Old Hash: %s\n New Hash:%s\n", storedHash, pagehash)
}
