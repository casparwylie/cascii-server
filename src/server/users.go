package main

import (
	"database/sql"
)

func MakeSessionKey() string {
	return GenerateUUID()
}

func CreateUser(db *sql.DB, email string, password string) error {
	_, err := db.Exec(
		"INSERT INTO users (email, password) VALUES (?, ?)",
		email,
		Hash(password),
	)
	return err
}

func GetUserById(db *sql.DB, id int) (string, error) {
	var email string
	err := db.QueryRow("SELECT email FROM users WHERE id = ?", id).Scan(&email)
	if err == nil || err == sql.ErrNoRows {
		return email, nil
	}
	return email, err
}

func UserExists(db *sql.DB, email string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT 1 FROM users WHERE email = ?", email).Scan(&exists)
	if err == nil || err == sql.ErrNoRows {
		return exists, nil
	}
	return exists, err
}

func Authenticate(db *sql.DB, email string, password string) (int, error) {
	userId := -1
	err := db.QueryRow(
		"SELECT id FROM users WHERE email = ? AND password = ?",
		email,
		Hash(password),
	).Scan(&userId)
	if err == nil || err == sql.ErrNoRows {
		return userId, nil
	}
	return userId, err
}

func CreateSession(db *sql.DB, userId int) (string, error) {
	key := MakeSessionKey()
	_, err := db.Exec(
		"INSERT INTO sessions (session_key, user_id) VALUES (?, ?)",
		key,
		userId,
	)
	if err == nil {
		return key, nil
	}
	return "", err
}

func DeleteSession(db *sql.DB, userId int) error {
	_, err := db.Exec("DELETE FROM sessions WHERE user_id = ?", userId)
	return err
}

func GetSessionUserId(db *sql.DB, key string) (int, error) {
	userId := -1
	err := db.QueryRow("SELECT user_id FROM sessions WHERE session_key = ?", key).Scan(&userId)
	if err == nil || err == sql.ErrNoRows {
		return userId, nil
	}
	return userId, err
}
