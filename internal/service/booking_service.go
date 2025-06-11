package service

import (
	"database/sql"

	"github.com/REmakerzz/dental-clinic-bot/internal/model"
	"github.com/REmakerzz/dental-clinic-bot/internal/repository"
)

type BookingService struct {
	db *sql.DB
}

func NewBookingService(db *sql.DB) *BookingService {
	return &BookingService{db: db}
}

func (s *BookingService) SaveBooking(booking *model.Booking) error {
	// Validate datetime before saving
	if err := repository.ValidateDateTime(s.db, booking.DateTime); err != nil {
		return err
	}
	return repository.SaveBooking(s.db, booking)
}

func (s *BookingService) DeleteBookingByID(id int) error {
	return repository.DeleteBookingByID(s.db, id)
}

func (s *BookingService) GetAllBookings() ([]*model.Booking, error) {
	return repository.GetAllBookings(s.db)
}

func (s *BookingService) GetBookingStats() (total int, today int, last7Days int, err error) {
	return repository.GetBookingStats(s.db)
}

// GetAvailableTimeSlots returns available time slots for a given date
func (s *BookingService) GetAvailableTimeSlots(date string) ([]string, error) {
	return repository.GetAvailableTimeSlots(s.db, date)
}

// IsDateTimeAvailable checks if the given datetime is available for booking
func (s *BookingService) IsDateTimeAvailable(datetime string) (bool, error) {
	return repository.IsDateTimeAvailable(s.db, datetime)
}

// ValidateDateTime checks if the datetime is in correct format and within working hours
func (s *BookingService) ValidateDateTime(datetime string) error {
	return repository.ValidateDateTime(s.db, datetime)
}
