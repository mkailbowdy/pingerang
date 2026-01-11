package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Site struct {
	ID      int
	Url     string
	Hash    string
	Created time.Time
}

type SiteModel struct {
	DB *sql.DB
}

func (m *SiteModel) Insert(url string, urlhash string, pagehash string) (int, error) {
	stmt := `INSERT INTO sites (url, created, urlhash, pagehash) VALUES (?, UTC_TIMESTAMP(), ?, ?)`

	result, err := m.DB.Exec(stmt, url, urlhash, pagehash)
	if err != nil {
		log.Fatal(err.Error())
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("DB insert complete!")
	return int(id), nil
}

func (m *SiteModel) Get(url string) (Site, error) {
	return Site{}, nil
}
