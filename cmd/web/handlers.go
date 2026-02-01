package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/mkailbowdy/internal/models"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, "home.tmpl.html", data)
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

	data := app.newTemplateData(r)
	data.Sites = sites
	app.render(w, r, http.StatusOK, "dashboard.tmpl.html", data)
}

func (app *application) createUrl(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("Rendering create URL page.")
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, "create.tmpl.html", data)
}

func (app *application) createUrlPost(w http.ResponseWriter, r *http.Request) {
	url, selector := app.getUrlSelectorPostForm(w, r)
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

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (app *application) getAndComparePost(w http.ResponseWriter, r *http.Request) {
	url, _ := app.getUrlSelectorPostForm(w, r)
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
	files := []string{
		"ui/html/partials/row.tmpl.html",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = ts.ExecuteTemplate(w, "change", nil)
	if err != nil {
		app.serverError(w, r, err)
		return
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
		for _, s := range sites {
			_, pagehash := driveHash(s.Url, s.Selector)
			err = app.compareHashes(s.Url, pagehash)
		}
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
	// sendUpdateMail(s.Url)

	return nil
}
func (app *application) updateHashesPost(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fmt.Printf("Updating hashes for ID: %s\n", id)
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	url := r.PostForm.Get("url")
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

	files := []string{
		"ui/html/partials/row.tmpl.html",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = ts.ExecuteTemplate(w, "nochange", nil)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}
