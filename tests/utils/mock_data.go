package utils

import (
	"fmt"
	models "github.com/chrisdamba/spacetrouble/internal"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func CreateMockBookings(count int) []models.Booking {
	bookings := make([]models.Booking, count)
	destinations := createMockDestinations()
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < count; i++ {
		bookings[i] = models.Booking{
			ID:     uuid.New(),
			Status: models.StatusConfirmed,
			User: models.User{
				ID:        uuid.New(),
				FirstName: fmt.Sprintf("Test%d", i+1),
				LastName:  fmt.Sprintf("User%d", i+1),
				Gender:    "other",
				Birthday:  baseTime.AddDate(-25-i, 0, 0), // Different ages
			},
			Flight: models.Flight{
				ID:          uuid.New(),
				LaunchpadID: fmt.Sprintf("5e9e4502f5090995de566f8%d", i+1), // Realistic SpaceX launchpad IDs
				LaunchDate:  baseTime.AddDate(0, i+1, 0),                   // Different launch dates
				Destination: destinations[i%len(destinations)],             // Cycle through destinations
			},
			CreatedAt: baseTime.Add(time.Duration(i) * time.Hour), // Different creation times
		}
	}
	return bookings
}

func createMockDestinations() []models.Destination {
	return []models.Destination{
		{
			ID:   uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"),
			Name: "Mars",
		},
		{
			ID:   uuid.MustParse("b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22"),
			Name: "Moon",
		},
		{
			ID:   uuid.MustParse("c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33"),
			Name: "Pluto",
		},
		{
			ID:   uuid.MustParse("d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a44"),
			Name: "Asteroid Belt",
		},
		{
			ID:   uuid.MustParse("e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a55"),
			Name: "Europa",
		},
		{
			ID:   uuid.MustParse("f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a66"),
			Name: "Titan",
		},
		{
			ID:   uuid.MustParse("70eebc99-9c0b-4ef8-bb6d-6bb9bd380a77"),
			Name: "Ganymede",
		},
	}
}

func CreateMockBookingRequest() *models.BookingRequest {
	return &models.BookingRequest{
		FirstName:     "Test",
		LastName:      "User",
		Gender:        "other",
		Birthday:      time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		LaunchpadID:   "5e9e4502f5090995de566f86",
		DestinationID: uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"), // Mars
		LaunchDate:    time.Now().AddDate(0, 1, 0),                            // One month from now
	}
}

func CreateMockBooking(id uuid.UUID) *models.Booking {
	if id == uuid.Nil {
		id = uuid.New()
	}

	return &models.Booking{
		ID:     id,
		Status: models.StatusConfirmed,
		User: models.User{
			ID:        uuid.New(),
			FirstName: "Test",
			LastName:  "User",
			Gender:    "other",
			Birthday:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		Flight: models.Flight{
			ID:          uuid.New(),
			LaunchpadID: "5e9e4502f5090995de566f86",
			LaunchDate:  time.Now().AddDate(0, 1, 0),
			Destination: models.Destination{
				ID:   uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"),
				Name: "Mars",
			},
		},
		CreatedAt: time.Now().UTC(),
	}
}

func CreateMockFlight() *models.Flight {
	return &models.Flight{
		ID:          uuid.New(),
		LaunchpadID: "5e9e4502f5090995de566f86",
		LaunchDate:  time.Now().AddDate(0, 1, 0),
		Destination: models.Destination{
			ID:   uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"),
			Name: "Mars",
		},
	}
}

func CreateMockBookingsResponse(count int) *models.AllBookingsResponse {
	bookings := CreateMockBookings(count)
	response := &models.AllBookingsResponse{
		Bookings: make([]models.BookingResponse, len(bookings)),
		Limit:    count,
		Cursor:   "mock-cursor",
	}

	for i, booking := range bookings {
		response.Bookings[i] = models.BookingResponse{Booking: booking}
	}

	return response
}

func BookingsEqual(t *testing.T, expected, actual *models.Booking) {
	t.Helper()

	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Status, actual.Status)

	assert.Equal(t, expected.User.ID, actual.User.ID)
	assert.Equal(t, expected.User.FirstName, actual.User.FirstName)
	assert.Equal(t, expected.User.LastName, actual.User.LastName)
	assert.Equal(t, expected.User.Gender, actual.User.Gender)

	assert.Equal(t, expected.Flight.ID, actual.Flight.ID)
	assert.Equal(t, expected.Flight.LaunchpadID, actual.Flight.LaunchpadID)
	assert.Equal(t, expected.Flight.Destination.ID, actual.Flight.Destination.ID)
	assert.Equal(t, expected.Flight.Destination.Name, actual.Flight.Destination.Name)
}

func BookingSlicesEqual(t *testing.T, expected, actual []models.Booking) {
	t.Helper()

	assert.Equal(t, len(expected), len(actual))
	for i := range expected {
		BookingsEqual(t, &expected[i], &actual[i])
	}
}
