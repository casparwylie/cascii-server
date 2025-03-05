package main

import (
	"database/sql"

	"github.com/VividCortex/mysqlerr"
	"github.com/go-sql-driver/mysql"
)

type MutableDrawingRow struct {
	Id        int
	Name      string
	CreatedAt string
}

func CreateImmutableDrawing(db *sql.DB, data string) (string, error) {
	hash := Hash(data)
	var err error
	for i := 5; i < 10; i++ {
		shortKey := hash[:i]
		_, err := db.Exec(
			"INSERT INTO immutable_drawings (short_key, hash, data) VALUES (?, ?, ?)",
			shortKey,
			hash,
			data,
		)
		if err == nil {
			return shortKey, nil
		}
		if driverErr, ok := err.(*mysql.MySQLError); ok {
			if driverErr.Number == mysqlerr.ER_DUP_ENTRY {
				// Our short key is already used, we try and resolve this...
				var existingHash string
				err = db.QueryRow(
					"SELECT hash FROM immutable_drawings WHERE short_key = ?", shortKey,
				).Scan(&existingHash)
				// The existing drawing is the same as the requested one, so we can
				// just use that one.
				if hash == existingHash {
					return shortKey, nil
				}
				// The drawings are different - the conflict is just bad luck!
				// We try again with a longer short key.
				continue
			}
		}
		// The error is unrelated to duplicates, so we can't fix it.
		return "", err
	}
	// Surely impossible, but there was no conflict resolution after 10 characters.
	return "", err
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

func UpdateMutableDrawing(db *sql.DB, drawingId int, data string, name string, userId int) (bool, error) {
	res, err := db.Exec(
		`UPDATE mutable_drawings SET
		data = COALESCE(NULLIF(?, ''), data),
		name = COALESCE(NULLIF(?, ''), name)
		WHERE id = ? AND user_id = ?`,
		data,
		name,
		drawingId,
		userId,
	)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

func GetMutableDrawing(db *sql.DB, drawingId int, userId int) (string, string, string, error) {
	var data string
	var createdAt string
	var name string
	err := db.QueryRow(
		"SELECT name, data, created_at FROM mutable_drawings WHERE id = ? AND user_id = ?", drawingId, userId,
	).Scan(&name, &data, &createdAt)
	if err == nil || err == sql.ErrNoRows {
		return name, data, createdAt, nil
	}
	return name, data, createdAt, err
}

func DeleteMutableDrawing(db *sql.DB, drawingId int, userId int) (bool, error) {
	res, err := db.Exec(
		"DELETE FROM mutable_drawings WHERE id = ? AND user_id = ?",
		drawingId,
		userId,
	)
	if err != nil {
		return false, err
	}
	deleted, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return deleted == 1, nil
}

func ListMutableDrawings(db *sql.DB, userId int) ([]MutableDrawingRow, error) {
	var results []MutableDrawingRow
	rows, err := db.Query(
		"SELECT id, name, created_at FROM mutable_drawings WHERE user_id = ? ORDER BY created_at DESC LIMIT 100",
		userId,
	)
	defer rows.Close()
	if err != nil {
		return results, err
	}
	var (
		id        int
		name      string
		createdAt string
	)
	for rows.Next() {
		err := rows.Scan(&id, &name, &createdAt)
		if err != nil {
			return results, err
		}
		results = append(results, MutableDrawingRow{id, name, createdAt})
	}
	return results, nil
}
