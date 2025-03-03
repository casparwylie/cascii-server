package main

import (
	"database/sql"
)

func MakeShortKey(seed string) string {
	return Hash(seed)[:6]
}

func CreateImmutableDrawing(db *sql.DB, data string) (string, error) {
	shortKey := MakeShortKey(data)
	_, err := db.Exec(
		"INSERT INTO immutable_drawings (short_key, data) VALUES (?, ?)",
		shortKey,
		data,
	)
	return shortKey, err
}

func GetImmutableDrawing(db *sql.DB, shortKey string) (string, string, error) {
	var data string
	var createdAt string
	err := db.QueryRow("SELECT data, created_at FROM immutable_drawings WHERE short_key = ?", shortKey).Scan(&data, &createdAt)
	if err == nil || err == sql.ErrNoRows {
		return data, createdAt, nil
	}
	return data, createdAt, err
}

func CreateMutableDrawing(db *sql.DB, data string, name string, userId int) (int, error) {
	res, err := db.Exec(
		"INSERT INTO mutable_drawings (user_id, name, data) VALUES (?, ?, ?)",
		userId,
		name,
		data,
	)
	if err != nil {
		return -1, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return int(id), err
}

func UpdateMutableDrawing(db *sql.DB, drawingId int, data string, name string, userId int) error {
	_, err := db.Exec(
		"UPDATE mutable_drawings SET data = ? AND name = ? WHERE id = ?",
		data,
		name,
		drawingId,
	)
	return err
}

func GetMutableDrawing(db *sql.DB, id int) (int, string, string, string, error) {
	var userId int
	var data string
	var createdAt string
	var name string
	err := db.QueryRow(
		"SELECT user_id, name, data, created_at FROM mutable_drawings WHERE id = ?", id,
	).Scan(&userId, &name, &data, &createdAt)
	if err == nil || err == sql.ErrNoRows {
		return userId, data, createdAt, name, nil
	}
	return userId, data, createdAt, name, err
}
