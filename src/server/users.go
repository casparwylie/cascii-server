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
	if err == nil || err == sql.ErrNoRows {
		return email
	}
	panic(err)
}

func UserExists(db *sql.DB, email string) bool {
	var exists bool
	err := db.QueryRow("SELECT 1 FROM users WHERE email = ?", email).Scan(&exists)
	if err == nil || err == sql.ErrNoRows {
		return exists
	}
	panic(err)
}

func Authenticate(db *sql.DB, email string, password string) int {
	userId := -1
	err := db.QueryRow(
		"SELECT id FROM users WHERE email = ? AND password = ?",
		email,
		HashPassword(password),
	).Scan(&userId)
	if err == nil || err == sql.ErrNoRows {
		return userId
	}
	panic(err)
}

func MakeSessionKey() string {
	return GenerateUUID()
}

func CreateSession(db *sql.DB, userId int) string {
	key := MakeSessionKey()
	_, err := db.Exec(
		"INSERT INTO sessions (session_key, user_id) VALUES (?, ?)",
		key,
		userId,
	)
	if err != nil {
		panic(err)
	}
	return key
}

func GetSessionUserId(db *sql.DB, key string) int {
	userId := -1
	err := db.QueryRow("SELECT user_id FROM sessions WHERE session_key = ?", key).Scan(&userId)
	if err == nil || err == sql.ErrNoRows {
		return userId
	}
	panic(err)
}
