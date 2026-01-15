package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mkailbowdy/internal/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from Pingerang"))
}

func (app *application) urlCreatePost(w http.ResponseWriter, r *http.Request) {
	url, selector := urlSelectorPostForm(r)
	fmt.Printf("%s and %s\n", url, selector)
	urlhash, pagehash := driveHash(url, selector)

	_, err := app.sites.Insert(url, urlhash, pagehash, selector)
	if err != nil {
		fmt.Printf("%s", err.Error())
		return
	}
}

func (app *application) urlComparePost(w http.ResponseWriter, r *http.Request) {
	url := urlPostForm(r)
	s, err := app.sites.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	urlhash, pagehash := driveHash(s.Url, s.Selector)

	err = app.compare(url, urlhash, pagehash)
	if err != nil {
		log.Fatal(err)
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
			log.Fatal(err)
			return
		}
		for i, s := range sites {
			fmt.Printf("Check #%d\n", i)
			urlhash, pagehash := driveHash(s.Url, s.Selector)
			err = app.compare(s.Url, urlhash, pagehash)
		}
	}
}

func (app *application) compare(url string, urlhash string, pagehash string) error {
	fmt.Printf("Now checking: %s\n", url)
	s, err := app.sites.Get(url)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			fmt.Printf("This url has not been registered.\n")
			return err
		} else {
			fmt.Printf("Error: %s", err.Error())
		}
	}
	if s.Pagehash == pagehash {
		fmt.Printf("No changes on this page.\n")
		return nil
	}
	fmt.Printf("The page has changed!\n")
	err = app.sites.Update(urlhash, pagehash)
	if err != nil {
		fmt.Printf("%s", err.Error())
		return err
	}
	fmt.Printf("-Old Hash: %s\n-New Hash: %s\n", s.Pagehash, pagehash)
	return nil
}
