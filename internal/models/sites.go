package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type SiteModelInterface interface {
	Insert(url, urlhash, pagehash, selector string) (int, error)
	Get(url string) (Site, error)
	GetAll() ([]Site, error)
	MarkAsChanged(urlhash string) error
	Update(urlhash, pagehash string) error
	Delete(id int) error
}

type Site struct {
	ID        int
	Url       string
	Created   time.Time
	Urlhash   string
	Pagehash  string
	Selector  string
	Changed   bool
	UpdatedAt time.Time
}

type SiteModel struct {
	DB *sql.DB
}

func (m *SiteModel) Insert(url, urlhash, pagehash, selector string) (int, error) {
	stmt := `INSERT INTO sites (url, created, urlhash, pagehash, selector) VALUES (?, UTC_TIMESTAMP(), ?, ?, ?)`

	result, err := m.DB.Exec(stmt, url, urlhash, pagehash, selector)
	if err != nil {
		fmt.Printf("DB Insert Error: %s", err.Error())
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		fmt.Printf("Last Insert ID Error: %s", err.Error())
		return 0, err
	}
	fmt.Println("DB insert complete!")
	return int(id), nil
}

func (m *SiteModel) Get(url string) (Site, error) {
	stmt := `SELECT id, url, created, urlhash, pagehash, selector, changed, updated_at FROM sites WHERE url = ?`
	row := m.DB.QueryRow(stmt, url)
	var s Site
	err := row.Scan(&s.ID, &s.Url, &s.Created, &s.Urlhash, &s.Pagehash, &s.Selector, &s.Changed, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Site{}, ErrNoRecord
		} else {
			return Site{}, err
		}
	}
	return s, err
}

func (m *SiteModel) MarkAsChanged(urlhash string) error {
	stmt := `UPDATE sites SET changed = ? WHERE urlhash = ?`
	_, err := m.DB.Exec(stmt, 1, urlhash)
	if err != nil {
		fmt.Printf("MarkAsChanged Error: %s", err.Error())
		return err
	}
	fmt.Println("Site is now marked as Changed.")
	return nil
}

func (m *SiteModel) Update(urlhash, pagehash string) error {
	stmt := `UPDATE sites SET pagehash = ?, changed = ? WHERE urlhash = ?`
	_, err := m.DB.Exec(stmt, pagehash, 0, urlhash)
	if err != nil {
		fmt.Printf("DB Update Error: %s", err.Error())
		return err
	}
	fmt.Println("Site is now marked as No Changes and pagehash has been updated.")
	return nil
}

func (m *SiteModel) Delete(id int) error {
	stmt := `DELETE FROM sites WHERE id = ?`
	_, err := m.DB.Exec(stmt, id)
	if err != nil {
		return err
	}
	fmt.Println("Url site removed successfully")
	return nil
}

func (m *SiteModel) GetAll() ([]Site, error) {
	stmt := `SELECT id, url, created, urlhash, pagehash, selector, changed, updated_at FROM sites`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		fmt.Printf("GetAll errorError: %s", err.Error())
		return nil, err
	}
	defer rows.Close()

	var sites []Site
	for rows.Next() {
		var s Site
		err = rows.Scan(&s.ID, &s.Url, &s.Created, &s.Urlhash, &s.Pagehash, &s.Selector, &s.Changed, &s.UpdatedAt)
		if err != nil {
			fmt.Printf("GetAll Error: %s", err.Error())
			return nil, err
		}
		sites = append(sites, s)
	}
	return sites, nil
}
