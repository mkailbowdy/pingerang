package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mkailbowdy/internal/models"
)

type application struct {
	sites *models.SiteModel
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from Pingerang"))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", home)
	url := "https://p-bandai.jp/item/item-1000243597/"
	dsn := flag.String("dsn", "web:Soul2001@/pingerang?parseTime=true", "MySQL data source name")
	flag.Parse()
	db, err := openDB(*dsn)
	if err != nil {
		fmt.Printf("error opening database pool: %s", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	app := &application{
		sites: &models.SiteModel{DB: db},
	}

	urlhash, pagehash, err := drive(url)
	if err != nil {
		fmt.Printf("There was an error: %q", err)
		os.Exit(1)
	}

	_, err = app.sites.Insert(url, urlhash, pagehash)
	if err != nil {
		fmt.Printf("There was an error: %q", err)
		os.Exit(1)
	}

	err = http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}

func drive(url string) (string, string, error) {
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
		fmt.Println("drive function: stop 1")
		log.Fatal(err)
	}
	hash := sha256.New()
	hash.Write([]byte(url))
	urlhash := fmt.Sprintf("%x", hash.Sum(nil))

	hash = sha256.New()
	hash.Write([]byte(html))
	pagehash := fmt.Sprintf("%x", hash.Sum(nil))
	return urlhash, pagehash, nil
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
