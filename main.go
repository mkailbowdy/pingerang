package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"flag"
	"fmt"
	"hash"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {

	dsn := flag.String("dsn", "web:Soul2001@/pingerang?parseTime=true", "MySQL data source name")
	flag.Parse()
	db, err := openDB(*dsn)
	if err != nil {
		fmt.Println("error opening database pool")
		os.Exit(1)
	}
	defer db.Close()
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

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
