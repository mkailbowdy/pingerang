package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

// UserModel struct wraps a database connection pool.
type UserModel struct {
	DB *sql.DB
}

// Add a new record to Users table in database
func (m *UserModel) Insert(name, email, password string) error {
	return nil
}

// Verify if the user exists with the provided email and password. Return userId if exists.
func (m *UserModel) Authenticate() (int, error) {
	return 0, nil
}

// Use Exists method to check if a user exists with a specific ID
func (m *UserModel) Exists(id int) (bool, error) {
	return false, nil
}
