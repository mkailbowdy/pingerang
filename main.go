package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"hash"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	url := "https://p-bandai.jp/item/item-1000241724/"
	h := drive(url)
	sites := make(map[string][]byte)
	sites[url] = h.Sum(nil)

	fmt.Printf("%x\n", sites[url])
}

func drive(url string) hash.Hash {
	var html string
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	err := chromedp.Run(
		ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(3*time.Second),
		chromedp.OuterHTML("body", &html),
	)
	if err != nil {
		fmt.Println("stop 1")
		log.Fatal(err)
	}
	h := sha256.New()
	h.Write([]byte(html))
	return h
}
