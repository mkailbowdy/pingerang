package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

func driveHash(url string) (string, string) {
	var html string
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	err := chromedp.Run(
		ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(5*time.Second),
		chromedp.OuterHTML("body", &html),
	)
	if err != nil {
		fmt.Println("driveHash function: stop 1")
		log.Fatal(err)
	}
	hash := sha256.New()
	hash.Write([]byte(url))
	urlhash := fmt.Sprintf("%x", hash.Sum(nil))

	hash = sha256.New()
	hash.Write([]byte(html))
	pagehash := fmt.Sprintf("%x", hash.Sum(nil))

	return urlhash, pagehash
}
