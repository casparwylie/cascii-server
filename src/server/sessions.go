package main

import "database/sql"

func MakeSessionKey() string {
	return GenerateUUID()
}

func CreateSession(db *sql.DB, userId int) string {
	key := MakeSessionKey()
	_, err := db.Exec(
		"INSERT INTO sessions (key, user_id) VALUES (?, ?)",
		key,
		userId,
	)
	if err != nil {
		panic(err)
	}
	return key
}

func GetSessionUserId(db *sql.DB, key string) int {
	var userId int
	err := db.QueryRow("SELECT user_id FROM sessions WHERE key = ?", key).Scan(&userId)
	switch err {
	case sql.ErrNoRows:
		return -1
	case nil:
		return userId
	default:
		panic(err)
	}
}
