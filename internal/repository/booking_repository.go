package repository

import (
    "database/sql"
    "github.com/REmakerzz/dental-clinic-bot/internal/model"
    "log"
	"time"
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