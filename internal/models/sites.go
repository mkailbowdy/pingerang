package models

import (
	"database/sql"
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

func (m *SiteModel) Insert(url string, hash string) (int, error) {
	return 0, nil
}

func (m *SiteModel) Get(url string) (Site, error) {
	return Site{}, nil
}
