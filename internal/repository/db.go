package repository

import (
    "database/sql"
    "log"

    _ "modernc.org/sqlite"
)

func InitDB() *sql.DB {
    db, err := sql.Open("sqlite", "clinic.db")
    if err != nil {
        log.Fatal(err)
    }

   createTableSQL := `CREATE TABLE IF NOT EXISTS bookings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    phone TEXT,
    service TEXT,
    datetime TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`

    _, err = db.Exec(createTableSQL)
    if err != nil {
        log.Fatalf("Failed to create table: %v", err)
    }

    return db
}