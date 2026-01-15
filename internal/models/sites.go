package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Site struct {
	ID       int
	Url      string
	Created  time.Time
	Urlhash  string
	Pagehash string
	Selector string
}

type SiteModel struct {
	DB *sql.DB
}

func (m *SiteModel) Insert(url, urlhash, pagehash, selector string) (int, error) {
	stmt := `INSERT INTO sites (url, created, urlhash, pagehash, selector) VALUES (?, UTC_TIMESTAMP(), ?, ?, ?)`

	result, err := m.DB.Exec(stmt, url, urlhash, pagehash, selector)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	fmt.Println("DB insert complete!")
	return int(id), nil
}

func (m *SiteModel) Get(url string) (Site, error) {
	stmt := `SELECT id, url, created, urlhash, pagehash, selector FROM sites WHERE url = ?`
	row := m.DB.QueryRow(stmt, url)
	var s Site
	err := row.Scan(&s.ID, &s.Url, &s.Created, &s.Urlhash, &s.Pagehash, &s.Selector)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Site{}, ErrNoRecord
		} else {
			return Site{}, err
		}
	}
	return s, err
}

func (m *SiteModel) Update(urlhash, pagehash string) error {
	stmt := `UPDATE sites SET pagehash = ? WHERE urlhash = ?`
	_, err := m.DB.Exec(stmt, pagehash, urlhash)
	if err != nil {
		fmt.Printf("%s", err.Error())
		return err
	}
	fmt.Println("Record updated")
	return nil
}

func (m *SiteModel) GetAll() ([]Site, error) {
	stmt := `SELECT id, url, created, urlhash, pagehash, selector FROM sites`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []Site
	for rows.Next() {
		var s Site
		err = rows.Scan(&s.ID, &s.Url, &s.Created, &s.Urlhash, &s.Pagehash, &s.Selector)
		if err != nil {
			return nil, err
		}
		sites = append(sites, s)
	}
	return sites, nil
}
