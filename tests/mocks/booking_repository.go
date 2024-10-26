package mocks

import (
	"context"
	models "github.com/chrisdamba/spacetrouble/internal"
	"github.com/stretchr/testify/mock"
	"time"
)

type MockBookingRepository struct {
	mock.Mock
}

func (m *MockBookingRepository) CreateBooking(ctx context.Context, booking *models.Booking) (*models.Booking, error) {
	args := m.Called(ctx, booking)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Booking), args.Error(1)
}

func (m *MockBookingRepository) GetDestinationById(ctx context.Context, id string) (*models.Destination, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Destination), args.Error(1)
}

func (m *MockBookingRepository) GetFlights(ctx context.Context, filters map[string]interface{}) ([]models.Flight, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Flight), args.Error(1)
}

func (m *MockBookingRepository) IsLaunchPadWeekAvailable(ctx context.Context, launchpadId, destinationId string, t time.Time) (bool, error) {
	args := m.Called(ctx, launchpadId, destinationId, t)
	return args.Bool(0), args.Error(1)
}

func (m *MockBookingRepository) GetBookingsPaginated(ctx context.Context, afterCursor string, limit int) ([]models.Booking, string, error) {
	args := m.Called(ctx, afterCursor, limit)
	return args.Get(0).([]models.Booking), args.String(1), args.Error(2)
}
