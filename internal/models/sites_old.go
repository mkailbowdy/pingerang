package models

//
//import (
//	"database/sql"
//	"errors"
//	"fmt"
//	"time"
//)
//
//type Site struct {
//	ID      int
//	Url     string
//	Hash    string
//	Created time.Time
//}
//
//type SiteModel struct {
//	DB *sql.DB
//}
//
//func (m *SiteModel) Insert(url string, urlhash string, pagehash string) (int, error) {
//	stmt := `INSERT INTO sites (url, created, urlhash, pagehash) VALUES (?, UTC_TIMESTAMP(), ?, ?)`
//
//	result, err := m.DB.Exec(stmt, url, urlhash, pagehash)
//	if err != nil {
//		return 0, err
//	}
//	id, err := result.LastInsertId()
//	if err != nil {
//		return 0, err
//	}
//	fmt.Println("DB insert complete!")
//	return int(id), nil
//}
//
//func (m *SiteModel) GetPageHash(urlhash string) (string, error) {
//	stmt := `SELECT pagehash FROM sites WHERE urlhash = ?`
//	row := m.DB.QueryRow(stmt, urlhash)
//	var storedHash string
//	err := row.Scan(&storedHash)
//	if err != nil {
//		if errors.Is(err, sql.ErrNoRows) {
//			return "", ErrNoRecord
//		} else {
//			return "", err
//		}
//	}
//	return storedHash, err
//}
//
//func (m *SiteModel) Update(urlhash, pagehash string) error {
//	stmt := `UPDATE sites SET pagehash = ? WHERE urlhash = ?`
//	_, err := m.DB.Exec(stmt, urlhash, pagehash)
//	if err != nil {
//		fmt.Printf("%s", err.Error())
//		return err
//	}
//	return nil
//}
//
//func (m *SiteModel) Urls() ([]string, error) {
//	stmt := `SELECT urlhash FROM sites`
//	rows, err := m.DB.Query(stmt)
//	if err != nil {
//		return nil, err
//	}
//	defer rows.Close()
//
//	var urlhashes []string
//	for rows.Next() {
//		var urlhash string
//		err = rows.Scan(&urlhash)
//		if err != nil {
//			return nil, err
//		}
//		urlhashes = append(urlhashes, urlhash)
//	}
//	return urlhashes, nil
//}
