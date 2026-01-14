package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"

	"github.com/chromedp/chromedp"
)

func urlPostForm(r *http.Request) (string, string) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err.Error())
	}
	url := r.PostForm.Get("url")
	selector := r.PostForm.Get("selector")
	return url, selector
}

func driveHash(url, selector string) (string, string) {

	var html string
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.Flag("headless", false),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	err := chromedp.Run(
		taskCtx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(selector, chromedp.ByQuery),
		chromedp.OuterHTML(selector, &html, chromedp.ByQuery),
	)
	if err != nil {
		fmt.Println("driveHash function: stop 1")
		log.Fatal(err)
	}
	fmt.Println("HTML:\n", html)

	hash := sha256.New()
	hash.Write([]byte(url))
	urlhash := fmt.Sprintf("%x", hash.Sum(nil))
	fmt.Printf("urlhash: %s\n", urlhash)

	hash.Reset()
	hash.Write([]byte(html))
	pagehash := fmt.Sprintf("%x", hash.Sum(nil))
	fmt.Printf("pagehash: %s\n", pagehash)

	return urlhash, pagehash
}
