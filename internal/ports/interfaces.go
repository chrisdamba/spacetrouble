package ports

import (
	"context"
	models "github.com/chrisdamba/spacetrouble/internal"
)

type BookingRepository interface {
	CreateBooking(ctx context.Context, booking *models.Booking) (*models.Booking, error)
	GetBookingsPaginated(ctx context.Context, afterCursor string, limit int) ([]models.Booking, string, error)
	GetDestinationById(ctx context.Context, id string) (*models.Destination, error)
	GetFlights(ctx context.Context, filters map[string]interface{}) ([]models.Flight, error)
}
