package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/mkailbowdy/internal/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")
	w.Write([]byte("Hello from Pingerang"))
}

func (app *application) urlCreatePost(w http.ResponseWriter, r *http.Request) {
	url, selector := urlSelectorPostForm(r)
	urlhash, pagehash := driveHash(url, selector)
	app.logger.Info("Hashes created.", "urlhash", urlhash)
	if len(urlhash) == 0 || len(pagehash) == 0 {
		app.logger.Error("There's a problem with the css selector you're using. Please fix the syntax and try again.")
		return
	}
	_, err := app.sites.Insert(url, urlhash, pagehash, selector)
	if err != nil {
		app.logger.Error(err.Error())
		return
	}
}

func (app *application) urlComparePost(w http.ResponseWriter, r *http.Request) {
	url := urlPostForm(r)
	s, err := app.sites.Get(url)
	if err != nil {
		app.logger.Error(err.Error())
		return
	}
	urlhash, pagehash := driveHash(s.Url, s.Selector)

	err = app.compare(url, urlhash, pagehash)
	if err != nil {
		app.logger.Error(err.Error())
	}
}

func (app *application) urlCompareBackground() {
	// Once an hour
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		// Get all the urlhash from database and store in a []string
		sites, err := app.sites.GetAll()
		if err != nil {
			app.logger.Error(err.Error())
			return
		}
		fmt.Printf("Session started.\n")
		for i, s := range sites {
			fmt.Printf("Site %d\n", i+1)
			urlhash, pagehash := driveHash(s.Url, s.Selector)
			err = app.compare(s.Url, urlhash, pagehash)
		}
		fmt.Printf("Session complete.\n\n")
	}
}

func (app *application) compare(url string, urlhash string, pagehash string) error {
	s, err := app.sites.Get(url)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.logger.Error("This url has not been registered.\n")
			return err
		} else {
			app.logger.Error(err.Error())
			return err
		}
	}
	if s.Pagehash == pagehash {
		fmt.Printf("No changes on this page.\n")
		return nil
	}
	fmt.Printf("The page has changed!")
	sendUpdateMail(s.Url)
	err = app.sites.Update(urlhash, pagehash)
	if err != nil {
		app.logger.Error(err.Error())
		return err
	}
	return nil
}
