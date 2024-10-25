package repository_test

import (
	"context"
	"github.com/chrisdamba/spacetrouble/internal/repository"
	"github.com/google/uuid"
	"regexp"
	"testing"
	"time"

	models "github.com/chrisdamba/spacetrouble/internal"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateBooking(t *testing.T) {
	mockDb, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockDb.Close()

	repo := repository.NewBookingRepository(mockDb)

	booking := &models.Booking{
		ID: uuid.New(),
		User: models.User{
			ID:        uuid.New(),
			FirstName: "John",
			LastName:  "Doe",
			Gender:    "Male",
			Birthday:  time.Now().Add(-20 * 365 * 24 * time.Hour),
		},
		Flight: models.Flight{
			ID:          uuid.New(),
			LaunchpadID: "LP1",
			Destination: models.Destination{
				ID:   uuid.New(),
				Name: "Mars",
			},
			LaunchDate: time.Now().Add(24 * time.Hour),
		},
		Status:    models.StatusActive,
		CreatedAt: time.Now(),
	}

	// begin transaction
	mockDb.ExpectBegin()

	// mock createUserTx
	userID := uuid.New()
	booking.User.ID = userID
	userQuery := regexp.QuoteMeta(`
        INSERT INTO users (id, first_name, last_name, gender, birthday)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (id) DO NOTHING
    `)
	mockDb.ExpectExec(userQuery).
		WithArgs(userID, booking.User.FirstName, booking.User.LastName, booking.User.Gender, booking.User.Birthday).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// mock createFlightTx
	flightID := uuid.New()
	booking.Flight.ID = flightID
	flightQuery := regexp.QuoteMeta(`
        INSERT INTO flights (id, launchpad_id, destination_id, launch_date)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id) DO NOTHING
    `)
	mockDb.ExpectExec(flightQuery).
		WithArgs(flightID, booking.Flight.LaunchpadID, booking.Flight.Destination.ID, booking.Flight.LaunchDate).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// mock createBookingTx
	bookingID := uuid.New()
	booking.ID = bookingID
	booking.Status = models.StatusConfirmed
	booking.CreatedAt = time.Now().UTC()
	bookingQuery := regexp.QuoteMeta(`
        INSERT INTO bookings (id, user_id, flight_id, status, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `)
	mockDb.ExpectExec(bookingQuery).
		WithArgs(booking.ID, booking.User.ID, booking.Flight.ID, booking.Status, booking.CreatedAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// commit transaction
	mockDb.ExpectCommit()

	createdBooking, err := repo.CreateBooking(context.Background(), booking)
	require.NoError(t, err)
	assert.Equal(t, booking.ID, createdBooking.ID)
	assert.Equal(t, booking.User.FirstName, createdBooking.User.FirstName)
	assert.Equal(t, booking.Flight.LaunchpadID, createdBooking.Flight.LaunchpadID)

	err = mockDb.ExpectationsWereMet()
	require.NoError(t, err)
}
