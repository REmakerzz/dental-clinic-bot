package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/REmakerzz/dental-clinic-bot/internal/model"
)

func SaveBooking(db *sql.DB, booking *model.Booking) error {
	stmt, err := db.Prepare(`INSERT INTO bookings (name, phone, service, datetime) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(booking.Name, booking.Phone, booking.Service, booking.DateTime)
	if err != nil {
		return err
	}

	log.Printf("Saved booking: %+v", booking)
	return nil
}

func GetAllBookings(db *sql.DB) ([]*model.Booking, error) {
	rows, err := db.Query(`SELECT id, name, phone, service, datetime FROM bookings ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []*model.Booking
	for rows.Next() {
		var b model.Booking
		err := rows.Scan(&b.ID, &b.Name, &b.Phone, &b.Service, &b.DateTime)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, &b)
	}

	return bookings, nil
}

func GetBookingStats(db *sql.DB) (total int, today int, last7Days int, err error) {
	err = db.QueryRow(`SELECT COUNT(*) FROM bookings`).Scan(&total)
	if err != nil {
		return
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour)

	err = db.QueryRow(`SELECT COUNT(*) FROM bookings WHERE created_at >= ? AND created_at < ?`,
		todayStart.Format("2006-01-02 15:04:05"), todayEnd.Format("2006-01-02 15:04:05")).Scan(&today)
	if err != nil {
		return
	}

	last7DaysStart := now.AddDate(0, 0, -7)
	err = db.QueryRow(`SELECT COUNT(*) FROM bookings WHERE created_at >= ?`,
		last7DaysStart.Format("2006-01-02 15:04:05")).Scan(&last7Days)
	if err != nil {
		return
	}

	return
}

func DeleteBookingByID(db *sql.DB, id int) error {
	stmt, err := db.Prepare(`DELETE FROM bookings WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// IsDateTimeAvailable checks if the given datetime is available for booking
func IsDateTimeAvailable(db *sql.DB, datetime string) (bool, error) {
	// First check if the time is within working hours
	t, err := time.Parse("2006-01-02 15:04", datetime)
	if err != nil {
		return false, fmt.Errorf("invalid datetime format: %w", err)
	}

	// Get working hours for the day
	var startTime, endTime string
	err = db.QueryRow(`
        SELECT start_time, end_time 
        FROM working_hours 
        WHERE day_of_week = ? AND is_working = 1`,
		t.Weekday()).Scan(&startTime, &endTime)
	if err != nil {
		return false, fmt.Errorf("failed to get working hours: %w", err)
	}

	// Parse working hours
	start, err := time.Parse("15:04", startTime)
	if err != nil {
		return false, fmt.Errorf("invalid start time format: %w", err)
	}
	end, err := time.Parse("15:04", endTime)
	if err != nil {
		return false, fmt.Errorf("invalid end time format: %w", err)
	}

	// Check if the time is within working hours
	bookingTime := time.Date(0, 0, 0, t.Hour(), t.Minute(), 0, 0, time.UTC)
	if bookingTime.Before(start) || bookingTime.After(end) {
		return false, nil
	}

	// Check if the time slot is already booked
	var exists bool
	err = db.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM bookings 
            WHERE datetime = ?
        )`, datetime).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check booking existence: %w", err)
	}

	return !exists, nil
}

// GetAvailableTimeSlots returns available time slots for a given date
func GetAvailableTimeSlots(db *sql.DB, date string) ([]string, error) {
	// Parse the date
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Get working hours for the day
	var startTime, endTime string
	err = db.QueryRow(`
        SELECT start_time, end_time 
        FROM working_hours 
        WHERE day_of_week = ? AND is_working = 1`,
		t.Weekday()).Scan(&startTime, &endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get working hours: %w", err)
	}

	// Parse working hours
	start, err := time.Parse("15:04", startTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format: %w", err)
	}
	end, err := time.Parse("15:04", endTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end time format: %w", err)
	}

	// Generate time slots (30-minute intervals)
	var slots []string
	current := time.Date(t.Year(), t.Month(), t.Day(), start.Hour(), start.Minute(), 0, 0, t.Location())
	slotEnd := time.Date(t.Year(), t.Month(), t.Day(), end.Hour(), end.Minute(), 0, 0, t.Location())

	for current.Before(slotEnd) {
		slot := current.Format("2006-01-02 15:04")

		// Check if the slot is available
		available, err := IsDateTimeAvailable(db, slot)
		if err != nil {
			return nil, err
		}

		if available {
			slots = append(slots, slot)
		}

		current = current.Add(30 * time.Minute)
	}

	return slots, nil
}

// ValidateDateTime checks if the datetime is in correct format and within working hours
func ValidateDateTime(db *sql.DB, datetime string) error {
	// Check format
	_, err := time.Parse("2006-01-02 15:04", datetime)
	if err != nil {
		return fmt.Errorf("invalid datetime format. Please use YYYY-MM-DD HH:MM format")
	}

	// Check if available
	available, err := IsDateTimeAvailable(db, datetime)
	if err != nil {
		return err
	}
	if !available {
		return fmt.Errorf("this time slot is not available")
	}

	return nil
}
