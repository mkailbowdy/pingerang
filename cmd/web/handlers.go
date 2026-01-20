package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/mkailbowdy/internal/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

	// Include the navigation partial in the template files.
	files := []string{
		"./ui/html/base.tmpl.html",
		"./ui/html/partials/nav.tmpl.html",
		"./ui/html/partials/form.tmpl.html",
		"./ui/html/pages/home.tmpl.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = ts.ExecuteTemplate(w, "base", contact)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (app *application) createSitePost(w http.ResponseWriter, r *http.Request) {
	url, selector := getUrlSelectorPostForm(r)
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

func (app *application) getAndComparePost(w http.ResponseWriter, r *http.Request) {
	url := urlPostForm(r)
	s, err := app.sites.Get(url)
	if err != nil {
		app.logger.Error(err.Error())
		return
	}
	_, pagehash := driveHash(s.Url, s.Selector)

	err = app.compareHashes(url, pagehash)
	if err != nil {
		app.logger.Error(err.Error())
	}
}

func (app *application) getAllAndCompareRoutine() {
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
			_, pagehash := driveHash(s.Url, s.Selector)
			err = app.compareHashes(s.Url, pagehash)
		}
		fmt.Printf("Session complete.\n\n")
	}
}

func (app *application) compareHashes(url string, pagehash string) error {
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

	// Update this snippets Changed column to true (1 in mysql)
	err = app.sites.MarkAsChanged(s.Urlhash)
	if err != nil {
		app.logger.Error(err.Error())
	}
	sendUpdateMail(s.Url)

	return nil
}
func (app *application) updateHashesPost(w http.ResponseWriter, r *http.Request) {
	url := r.PostFormValue("url")
	fmt.Printf("%s", url)
	s, err := app.sites.Get(url)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	urlhash, pagehash := driveHash(s.Url, s.Selector)
	err = app.sites.Update(urlhash, pagehash)
	if err != nil {
		app.logger.Error(err.Error())
	}
}

func (app *application) dashboard(w http.ResponseWriter, r *http.Request) {
	sites, err := app.sites.GetAll()
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := &templateData{
		Sites: sites,
	}

	files := []string{
		"./ui/html/base.tmpl.html",
		"./ui/html/partials/nav.tmpl.html",
		"./ui/html/pages/dashboard.tmpl.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

type Contact struct {
	FirstName string
	LastName  string
	Email     string
}

var contact = Contact{
	FirstName: "Joe",
	LastName:  "Hisaishi",
	Email:     "studioghibli@gmail.com",
}

func (app *application) contact(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/base.tmpl.html",
		"./ui/html/partials/nav.tmpl.html",
		"./ui/html/partials/form.tmpl.html",
		"./ui/html/pages/contact.tmpl.html",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	ts.ExecuteTemplate(w, "base", contact)
}

func (app *application) viewForm(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/partials/form.tmpl.html",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	ts.ExecuteTemplate(w, "contact-view", contact)
}

func (app *application) editForm(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./ui/html/partials/form.tmpl.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	ts.ExecuteTemplate(w, "contact-edit", contact)
}
