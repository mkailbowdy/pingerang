package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
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
	// Initialize .Form because otherwise it will be nil when the template is parsed.
	data.Form = make(map[string]string)
	app.render(w, r, http.StatusOK, "create.tmpl.html", data)
}

// urlCreateForm represents the form data and validation errors for the form fields.
type urlCreateForm struct {
	Url         string
	Selector    string
	FieldErrors map[string]string
}

func (app *application) createUrlPost(w http.ResponseWriter, r *http.Request) {
	u, selector := app.getUrlSelectorPostForm(w, r)

	form := urlCreateForm{
		Url:         u,
		Selector:    selector,
		FieldErrors: map[string]string{},
	}
	fmt.Printf("Received URL: %s and Selector: %s\n", form.Url, form.Selector)
	validUrl, err := url.Parse(u)
	if err != nil {
		app.logger.Error(err.Error())
		return
	}

	if strings.TrimSpace(validUrl.String()) == "" {
		form.FieldErrors["url"] = "This field cannot be blank."
	} else if validUrl.Scheme == "" || validUrl.Host == "" {
		form.FieldErrors["url"] = "url must be a an absolute URL"
	} else if validUrl.Scheme != "http" && validUrl.Scheme != "https" {
		form.FieldErrors["url"] = "error: url must begin with http or https"
	}

	if strings.TrimSpace(form.Selector) == "" {
		form.FieldErrors["selector"] = "This field cannot be blank."
	}

	if len(form.FieldErrors) > 0 {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "create.tmpl.html", data)
		return
	}

	urlhash, pagehash := driveHash(form.Url, form.Selector)
	app.logger.Info("Hashes created.", "urlhash", urlhash)
	if len(urlhash) == 0 || len(pagehash) == 0 {
		app.logger.Error("There's a problem with the css selector you're using. Please fix the syntax and try again.")
		return
	}
	_, err = app.sites.Insert(form.Url, urlhash, pagehash, form.Selector)
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
}

func (app *application) getAllAndCompareRoutine() {
	// To Do: Run at the 50th minute of every hour.(e.g. 10:50, 11:50,...)
	// Once an hour
	ticker := time.NewTicker(20 * time.Minute)
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

	err = ts.ExecuteTemplate(w, "nochange", s)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}

func (app *application) showButton(w http.ResponseWriter, r *http.Request) {

	// look up the Site in the database
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	url := r.PostForm.Get("url")
	fmt.Printf("The URL is %s\n", url)
	s, err := app.sites.Get(url)
	fmt.Printf("Site values is: %v\n", s)
	files := []string{
		"ui/html/partials/row.tmpl.html",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	if s.Changed == false {
		err = ts.ExecuteTemplate(w, "nochange", s)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		return
	}
	err = ts.ExecuteTemplate(w, "change", s)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}
