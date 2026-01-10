package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mkailbowdy/internal/models"
)

type application struct {
	sites *models.SiteModel
}

func main() {
	url := "https://p-bandai.jp/item/item-1000241724/"
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

	err = drive(url)
	if err != nil {
		fmt.Printf("There was an error: %q", err)
		os.Exit(1)
	}

	_, err = app.sites.Insert(url)
	if err != nil {
		fmt.Printf("There was an error: %q", err)
		os.Exit(1)
	}

	return
}

func drive(url string) error {
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
		fmt.Println("drive function: stop 1")
		log.Fatal(err)
	}
	return nil
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
