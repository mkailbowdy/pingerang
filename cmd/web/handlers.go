package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/mkailbowdy/internal/models"
	"github.com/mkailbowdy/internal/validator"
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
	Url       string              `form:"url"`
	Selector  string              `form:"selector"`
	Validator validator.Validator `form:"-"`
}

func (app *application) createUrlPost(w http.ResponseWriter, r *http.Request) {
	form := app.getUrlSelectorPostForm(w, r)

	fmt.Printf("Received URL: %s and Selector: %s\n", form.Url, form.Selector)
	validUrl, err := url.Parse(form.Url)
	if err != nil {
		app.logger.Error(err.Error())
		return
	}

	// Check NotBlank and ValidUrl
	form.Validator.CheckField(validator.NotBlank(validUrl.String()), "url", "This field cannot be blank.")
	form.Validator.CheckField(validator.ValidUrl(*validUrl), "url", "This must be an absolute link that begins with http or https")
	form.Validator.CheckField(validator.NotBlank(form.Selector), "selector", "This field cannot be blank")
	if !form.Validator.Valid() {
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

	app.sessionManager.Put(r.Context(), "flash", "Url successfully created!")

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (app *application) getAndComparePost(s models.Site) {
	fmt.Printf("\ns.Url: %s\ns.Selector: %s\n", s.Url, s.Selector)
	_, pagehash := driveHash(s.Url, s.Selector)

	err := app.compareHashes(s.Url, pagehash)
	if err != nil {
		app.logger.Error(err.Error())
	}
}

func (app *application) removeSite(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if id <= 0 || err != nil {
		app.logger.Error(err.Error())
		return
	}
	err = app.sites.Delete(id)
	if err != nil {
		app.logger.Error(err.Error())
	}
}

func (app *application) getAllAndCompareRoutine() {
	// To Do: Run at the 50th minute of every hour.(e.g. 10:50, 11:50,...)
	// Once an hour
	ticker := time.NewTicker(5 * time.Minute)

	defer ticker.Stop()
	for range ticker.C {
		fmt.Printf("\nRunning: getAllAndCompareRoutine\n")
		// Get all the urlhash from database and store in a []string
		sites, err := app.sites.GetAll()
		if err != nil {
			app.logger.Error(err.Error())
			return
		}

		for _, s := range sites {
			go app.getAndComparePost(s)
		}
	}
	fmt.Printf("\ngetAllAndCompareRoutine finished.\n")
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
	fmt.Printf("\n%s\npagehash:%s\ns.Pagehash:%s\n\n", url, pagehash, s.Pagehash)
	if s.Pagehash == pagehash {
		fmt.Printf("No changes on this page.\n")
		return nil
	}
	fmt.Printf("The page has changed!")

	// Update this snippets Changed column to true (1 in mysql)
	err = app.sites.MarkAsChanged(s.Urlhash)
	if err != nil {
		fmt.Printf("compareHashes error: %s", err.Error())
		app.logger.Error(err.Error())
		return err
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
		return
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
	s, err := app.sites.Get(url)
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

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	app.render(w, r, http.StatusOK, "signup.tmpl.html", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank.")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank.")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email.")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank.")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be a minimum of 8 characters.")
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "signup.tmpl.html", data)
		return
	}
	// Placeholder Response
	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email Address already in use.")
			data := app.newTemplateData(r)
			data.Form = form
			fmt.Printf("%v", form)
			app.render(w, r, http.StatusUnprocessableEntity, "signup.tmpl.html", data)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, r, http.StatusOK, "login.tmpl.html", data)
}
func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form.CheckField(validator.NotBlank(form.Email), "email", " This field cannot be blank.")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email.")
	form.CheckField(validator.NotBlank(form.Password), "password", " This field cannot be blank.")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "login.tmpl.html", data)
		return
	}

	// Check if the credentials are valid. If not, add generic non-field error message and redisplay login screen

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldErrors("Email or password is incorrect.")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "login.tmpl.html", data)
		} else {
			app.serverError(w, r, err)
		}
		return
	}
	// Use the RenewToken() method on the current session to change the session
	// ID. It's good practice to generate a new session ID when the
	// authentication state or privilege levels change for the user (e.g. login
	// and logout operations).
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
	}
	// Add the ID of the current user to the session, so that they are now
	// 'logged in'.
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	// Use the RenewToken() method on the current session to change the session
	// ID again.`
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	// Remove the authenticatedUserID from the session data so that the user is
	// 'logged out'.
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	// Add a flash message to the session to confirm to the user that they've been
	// logged out.
	app.sessionManager.Put(r.Context(), "flash", "You've been successfully logged out.")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
