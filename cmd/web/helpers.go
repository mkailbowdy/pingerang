package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/go-playground/form/v4"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"sync"
	"time"
)

func (app *application) getUrlSelectorPostForm(w http.ResponseWriter, r *http.Request) urlCreateForm {
	var form urlCreateForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return urlCreateForm{}
	}
	// url := r.PostForm.Get("url")
	// selector := r.PostForm.Get("selector")
	// fmt.Printf("The URL is %s and the selector is %s\n", url, selector)
	return form
}

/*
func driveHash(url, selector string) (string, string) {

		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", false),
			chromedp.Flag("lang", "ja-JP"),
			chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
			chromedp.NoDefaultBrowserCheck,
			chromedp.Flag("disable-blink-features", "AutomationControlled"),
		)
		allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
		defer cancel()

		ctx, cancel := chromedp.NewContext(allocCtx)
		defer cancel()

		ctx, cancel = context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		var html string

		err := chromedp.Run(
			ctx,
			network.Enable(),
			network.SetExtraHTTPHeaders(network.Headers{
				"Accept-Language": "ja-JP,ja;q=0.9",
			}),
			chromedp.Navigate(url),
			// Wait for the specific element to appear in the DOM
			chromedp.Sleep(10*time.Second),
			chromedp.WaitVisible(selector, chromedp.ByQuery),
			chromedp.InnerHTML(selector, &html),
		)
		fmt.Printf("\nInnerHTML: %s\n", html)

		if err != nil {
			fmt.Printf("driveHash method: %s", err.Error())
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
*/
func driveHash(url, selector string) (string, string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	profileDir := fmt.Sprintf("/tmp/chrome-profile-%d", time.Now().UnixNano())

	defer os.RemoveAll(profileDir)
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),

		// --- Locale / Language ---
		chromedp.Flag("lang", "ja-JP"),
		chromedp.Flag("disable-features", "TranslateUI"),

		// --- Stealth ---
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("user-data-dir", profileDir),

		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.7680.177 Safari/537.36"),
		chromedp.NoDefaultBrowserCheck,
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	var html string

	// --- Track network activity ---
	var mu sync.Mutex
	inflight := 0

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		mu.Lock()
		defer mu.Unlock()

		switch ev.(type) {
		case *network.EventRequestWillBeSent:
			inflight++
		case *network.EventLoadingFinished, *network.EventLoadingFailed:
			if inflight > 0 {
				inflight--
			}
		}
	})

	err := chromedp.Run(
		ctx,
		network.Enable(),
		network.ClearBrowserCookies(),

		network.SetExtraHTTPHeaders(network.Headers{
			"Accept-Language":    "ja-JP,ja;q=0.9",
			"Sec-CH-UA":          `"Chromium";v="120", "Not=A?Brand";v="99"`,
			"Sec-CH-UA-Platform": `"Windows"`,
		}),

		emulation.SetTimezoneOverride("Asia/Tokyo"),

		// --- Fingerprint spoofing ---
		chromedp.Evaluate(`
			Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
			Object.defineProperty(navigator, 'languages', { get: () => ['ja-JP', 'ja'] });
			Object.defineProperty(navigator, 'language', { get: () => 'ja-JP' });
			Object.defineProperty(navigator, 'platform', { get: () => 'Win32' });
			Object.defineProperty(navigator, 'plugins', { get: () => [1,2,3,4,5] });
		`, nil),

		chromedp.Navigate(url),

		// --- Wait for element to exist ---
		chromedp.WaitVisible(selector, chromedp.ByQuery),
	)

	if err != nil {
		fmt.Printf("navigate error: %s\n", err)
		return "", ""
	}

	// --- Smart wait: ensure content is actually populated ---
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			for i := 0; i < 10; i++ {
				var ready bool
				err := chromedp.Evaluate(`(() => {
					const el = document.querySelector("`+selector+`");
					return el && el.innerText && el.innerText.length > 20;
				})()`, &ready).Do(ctx)

				if err == nil && ready {
					return nil
				}

				time.Sleep(time.Duration(1+r.Intn(2)) * time.Second)
			}
			return fmt.Errorf("content not ready")
		}),
	)

	if err != nil {
		fmt.Println("content readiness warning:", err)
		return "", ""
	}

	// --- Handle lazy loading / infinite scroll ---
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			var lastHeight, newHeight int64

			for i := 0; i < 5; i++ {
				// get current height
				if err := chromedp.Evaluate(`document.body.scrollHeight`, &lastHeight).Do(ctx); err != nil {
					return err
				}

				// scroll down
				if err := chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight)`, nil).Do(ctx); err != nil {
					return err
				}

				time.Sleep(time.Duration(2+r.Intn(3)) * time.Second)

				// get new height
				if err := chromedp.Evaluate(`document.body.scrollHeight`, &newHeight).Do(ctx); err != nil {
					return err
				}

				if newHeight == lastHeight {
					break // no more content
				}
			}
			return nil
		}),
	)

	if err != nil {
		fmt.Println("scroll warning:", err)
		return "", ""
	}

	// --- Wait for network to go idle ---
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			for i := 0; i < 10; i++ {
				mu.Lock()
				active := inflight
				mu.Unlock()

				if active == 0 {
					time.Sleep(1 * time.Second) // small buffer
					return nil
				}
				time.Sleep(500 * time.Millisecond)
			}
			return nil
		}),
	)

	// --- Extract final HTML ---
	err = chromedp.Run(ctx,
		chromedp.ScrollIntoView(selector),
		chromedp.InnerHTML(selector, &html),
	)

	if err != nil {
		fmt.Printf("extract error: %s\n", err)
		return "", ""
	}

	fmt.Printf("\nInnerHTML: %s\n", html)

	// --- Hash URL ---
	hash := sha256.New()
	hash.Write([]byte(url))
	urlhash := fmt.Sprintf("%x", hash.Sum(nil))

	// --- Hash content ---
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

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("The template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}

	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Write out the provided HTTP status code ('200 OK', '400 Bad Request' etc).
	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (app *application) decodePostForm(r *http.Request, destination any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = app.formDecoder.Decode(destination, r.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError

		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}
	}
	return err
}

func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}
	return isAuthenticated
}
