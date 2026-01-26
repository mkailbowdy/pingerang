package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"time"

	"github.com/chromedp/chromedp"
)

func getUrlSelectorPostForm(r *http.Request) (string, string) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err.Error())
	}
	url := r.PostForm.Get("url")
	fmt.Printf("%s\n", url)
	selector := r.PostForm.Get("selector")

	return url, selector
}
func urlPostForm(r *http.Request) string {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err.Error())
	}
	url := r.PostForm.Get("url")
	return url
}

func driveHash(url, selector string) (string, string) {

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var html string

	err := chromedp.Run(
		ctx,
		chromedp.Navigate(url),
		// Wait for the specific element to appear in the DOM
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.OuterHTML(selector, &html),
	)

	if err != nil {
		log.Println("something is wrong:", err)
		return "", ""
	}

	hash := sha256.New()
	hash.Write([]byte(url))
	urlhash := fmt.Sprintf("%x", hash.Sum(nil))

	hash.Reset()
	hash.Write([]byte(html))
	pagehash := fmt.Sprintf("%x", hash.Sum(nil))

	return urlhash, pagehash
}

func sendUpdateMail(url string) {
	// Set up authentication information.
	auth := smtp.PlainAuth("", "myhkail.mendoza@gmail.com", "kizrnvfnknzxolbn", "smtp.gmail.com")

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	to := []string{"kailphotoshoots@gmail.com"}
	msg := []byte("To: kailphotoshoots@gmail.com\r\n" +
		"Subject: A page you follow has changed!\r\n" +
		"\r\n" +
		"View the page that changed below! (:\r\n" +
		url)
	err := smtp.SendMail("smtp.gmail.com:587", auth, "myhkail.mendoza@gmail.com", to, msg)
	if err != nil {
		fmt.Printf("Error while trying to send email.")
		log.Fatal(err)
	}
}

// The serverError helper writes a log entry at Error level (including the request
// method and URI as attributes), then sends a generic 500 Internal Server Error
// response to the user.
func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// The clientError helper sends a specific status code and corresponding description
// to the user. We'll use this later in the book to send responses like 400 "Bad
// Request" when there's a problem with the request that the user sent.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData){
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("The template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}

	
	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil{
		app.serverError(w, r, err)
		return
	}

	// Write out the provided HTTP status code ('200 OK', '400 Bad Request' etc).
	w.WriteHeader(status)
	buf.WriteTo(w)
}