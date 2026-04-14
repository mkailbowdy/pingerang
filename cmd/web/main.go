package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"github.com/alexedwards/scs/mysqlstore" // New import
	"github.com/alexedwards/scs/v2"         // New import
	"github.com/chromedp/chromedp"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mkailbowdy/internal/models"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

type application struct {
	sites          models.SiteModelInterface
	users          models.UserModelInterface
	logger         *slog.Logger
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
	allocCtx       *context.Context
	wg             sync.WaitGroup
}

func main() {
	mysqlConnDsn := fmt.Sprintf("%s:%s@tcp(mysql:3306)/pingerang?parseTime=true",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
	)
	addr := flag.String("addr", ":4000", "HTTP network address")

	//dsn := flag.String("dsn", "web:Soul2001@/pingerang?parseTime=true", "MySQL data source name")
	// dsn := os.Getenv("DSN")

	flag.Parse()
	if os.Getenv("MYSQL_USER") == "" || os.Getenv("MYSQL_PASSWORD") == "" {
		mysqlConnDsn = "web:Soul2001@/pingerang?parseTime=true"
	}

	file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	logger := slog.New(slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	var db *sql.DB
	for i := 1; i <= 10; i++ {
		db, err = openDB(mysqlConnDsn)
		if err != nil {
			if i == 10 {
				fmt.Printf("Failed to open database: %s\n", err.Error())
				os.Exit(1)
			}
			fmt.Printf("Retrying connection to database")
			time.Sleep(1 * time.Second)
			continue
		}

		break
	}
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Secure = true

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("lang", "ja-JP"),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-gpu", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	app := &application{
		sites:          &models.SiteModel{DB: db},
		users:          &models.UserModel{DB: db},
		logger:         logger,
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
		allocCtx:       &allocCtx,
	}

	go app.getAllAndCompareRoutine()

	srv := &http.Server{
		Addr:         *addr,
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		IdleTimeout:  5 * time.Minute,
		ReadTimeout:  30 * time.Minute,
		WriteTimeout: 60 * time.Minute,
	}
	logger.Info("starting server", "addr", srv.Addr)
	fmt.Println("Listening and serving requests!")
	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
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
