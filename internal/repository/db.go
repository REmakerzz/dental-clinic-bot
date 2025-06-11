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

	// Create bookings table with unique constraint on datetime
	createBookingsTableSQL := `CREATE TABLE IF NOT EXISTS bookings (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        phone TEXT NOT NULL,
        service TEXT NOT NULL,
        datetime TEXT NOT NULL UNIQUE,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`

	// Create working hours table
	createWorkingHoursTableSQL := `CREATE TABLE IF NOT EXISTS working_hours (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        day_of_week INTEGER NOT NULL, -- 0-6 (Sunday-Saturday)
        start_time TEXT NOT NULL,     -- Format: "HH:MM"
        end_time TEXT NOT NULL,       -- Format: "HH:MM"
        is_working BOOLEAN NOT NULL DEFAULT 1
    );`

	// Create time slots table
	createTimeSlotsTableSQL := `CREATE TABLE IF NOT EXISTS time_slots (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        datetime TEXT NOT NULL UNIQUE,
        is_available BOOLEAN NOT NULL DEFAULT 1
    );`

	_, err = db.Exec(createBookingsTableSQL)
	if err != nil {
		log.Fatalf("Failed to create bookings table: %v", err)
	}

	_, err = db.Exec(createWorkingHoursTableSQL)
	if err != nil {
		log.Fatalf("Failed to create working hours table: %v", err)
	}

	_, err = db.Exec(createTimeSlotsTableSQL)
	if err != nil {
		log.Fatalf("Failed to create time slots table: %v", err)
	}

	// Insert default working hours if not exists
	insertWorkingHoursSQL := `INSERT OR IGNORE INTO working_hours (day_of_week, start_time, end_time) VALUES 
        (1, '09:00', '18:00'), -- Monday
        (2, '09:00', '18:00'), -- Tuesday
        (3, '09:00', '18:00'), -- Wednesday
        (4, '09:00', '18:00'), -- Thursday
        (5, '09:00', '18:00'), -- Friday
        (6, '10:00', '15:00'), -- Saturday
        (0, '00:00', '00:00'); -- Sunday (closed)`

	_, err = db.Exec(insertWorkingHoursSQL)
	if err != nil {
		log.Fatalf("Failed to insert default working hours: %v", err)
	}

	return db
}
