package models

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
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

func (m *SiteModel) Insert(url string) (int, error) {
	stmt := `INSERT INTO sites (url, hash, created) VALUES (?, ?, UTC_TIMESTAMP())`

	hash := sha256.New()
	hash.Write([]byte(url))
	hashHex := fmt.Sprintf("%x", hash.Sum(nil))
	result, err := m.DB.Exec(stmt, url, hashHex)
	if err != nil {
		fmt.Printf("There was an error: %q\n", err)
		return 0, nil
	}
	id, err := result.LastInsertId()
	if err != nil {
		fmt.Printf("There was an error: %q\n", err)
		return 0, nil
	}
	return int(id), nil
}

func (m *SiteModel) Get(url string) (Site, error) {
	return Site{}, nil
}
