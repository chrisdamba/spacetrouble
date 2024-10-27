package ports

import (
	"context"
	models "github.com/chrisdamba/spacetrouble/internal"
	"time"
)

type BookingRepository interface {
	CreateBooking(ctx context.Context, booking *models.Booking) (*models.Booking, error)
	GetBookingByID(ctx context.Context, id string) (*models.Booking, error)
	GetBookingsPaginated(ctx context.Context, afterCursor string, limit int) ([]models.Booking, string, error)
	GetDestinationById(ctx context.Context, id string) (*models.Destination, error)
	GetFlights(ctx context.Context, filters map[string]interface{}) ([]models.Flight, error)
	IsLaunchPadWeekAvailable(ctx context.Context, launchpadId, destinationId string,
		t time.Time) (bool, error)
	DeleteBooking(ctx context.Context, id string) error
}

type BookingService interface {
	CreateBooking(ctx context.Context, request *models.BookingRequest) (*models.Booking, error)
	AllBookings(ctx context.Context, req models.GetBookingsRequest) (*models.AllBookingsResponse, error)
	DeleteBooking(ctx context.Context, id string) error
}

type SpaceXClient interface {
	CheckLaunchConflict(ctx context.Context, launchpadID string, ts time.Time) (bool, error)
}
