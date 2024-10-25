package ports

import models "github.com/chrisdamba/spacetrouble/internal"

type BookingRepository interface {
	CreateBooking(booking *models.Booking) (*models.Booking, error)
}
