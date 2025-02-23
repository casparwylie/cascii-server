package main

import (
	"fmt"
	"os"

	// Docs: http://go-database-sql.org/accessing.html
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type DbFactory struct {
	db           *sql.DB
	maxConns     int
	maxIdleConns int
}

func (dbFactory *DbFactory) Get() *sql.DB {
	if dbFactory.db != nil {
		return dbFactory.db
	}
	connString := dbFactory.GetConnectionString()
	fmt.Println(connString)
	db, err := sql.Open("mysql", connString)
	dbFactory.db = db
	if err != nil {
		// TODO: consider better err handling
		panic(err)
	}
	db.SetMaxOpenConns(dbFactory.maxConns)
	db.SetMaxIdleConns(dbFactory.maxIdleConns)
	return db
}

func (dbFactory DbFactory) GetConnectionString() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
}
