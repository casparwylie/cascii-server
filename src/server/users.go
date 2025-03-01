package main

import (
	"database/sql"
)

func CreateUser(db *sql.DB, email string, password string) bool {
	_, err := db.Exec(
		"INSERT INTO users (email, password) VALUES (?, ?)",
		email,
		HashPassword(password),
	)
	return err == nil
}

func GetUserById(db *sql.DB, id int) string {
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

func Authenticate(db *sql.DB, email string, password string) int {
	userId := -1
	err := db.QueryRow(
		"SELECT id FROM users WHERE email = ? AND password = ?",
		email,
		HashPassword(password),
	).Scan(&userId)
	if err != nil {
		panic(err)
	}
	return userId
}
