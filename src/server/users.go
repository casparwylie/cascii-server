package main

import (
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
)

func HashPassword(password string) string {
	hash := sha512.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}

func CreateUser(db *sql.DB, email string, password string) bool {
	_, err := db.Exec(
		"INSERT INTO users (email, password) VALUES (?, ?)",
		email,
		HashPassword(password),
	)
	return err == nil
}

func GetUserById(db *sql.DB, id string) string {
	var email string
	err := db.QueryRow("SELECT email FROM users WHERE id = ?", id).Scan(&email)
	switch err {
	case sql.ErrNoRows:
		return ""
	case nil:
		return email
	default:
		panic(err)
	}
}

func UserExists(db *sql.DB, email string) bool {
	var exists bool
	err := db.QueryRow("SELECT 1 FROM users WHERE email = ?", email).Scan(&exists)
	switch err {
	case sql.ErrNoRows:
		return false
	case nil:
		return exists
	default:
		panic(err)
	}
}
