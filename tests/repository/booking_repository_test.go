package repository_test

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/chrisdamba/spacetrouble/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"regexp"
	"strings"
	"testing"
	"time"

	models "github.com/chrisdamba/spacetrouble/internal"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateBooking(t *testing.T) {
	mockDb, repo := setupMockDB(t)
	defer mockDb.Close()

	// create fixed UUIDs for testing
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	flightID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	bookingID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	destinationID := uuid.MustParse("00000000-0000-0000-0000-000000000004")

	booking := &models.Booking{
		ID: bookingID,
		User: models.User{
			ID:        userID,
			FirstName: "John",
			LastName:  "Doe",
			Gender:    "Male",
			Birthday:  time.Now().Add(-20 * 365 * 24 * time.Hour),
		},
		Flight: models.Flight{
			ID:          flightID,
			LaunchpadID: "LP1",
			Destination: models.Destination{
				ID:   destinationID,
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
	userQuery := regexp.QuoteMeta(`
        INSERT INTO users (id, first_name, last_name, gender, birthday)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (id) DO NOTHING
    `)
	mockDb.ExpectExec(userQuery).
		WithArgs(userID, booking.User.FirstName, booking.User.LastName, booking.User.Gender, booking.User.Birthday).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// mock createFlightTx
	flightQuery := regexp.QuoteMeta(`
        INSERT INTO flights (id, launchpad_id, destination_id, launch_date)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id) DO NOTHING
    `)
	mockDb.ExpectExec(flightQuery).
		WithArgs(flightID, booking.Flight.LaunchpadID, booking.Flight.Destination.ID, booking.Flight.LaunchDate).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// mock createBookingTx
	booking.Status = models.StatusConfirmed
	booking.CreatedAt = time.Now().UTC()
	bookingQuery := regexp.QuoteMeta(`
        INSERT INTO bookings (id, user_id, flight_id, status, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `)
	mockDb.ExpectExec(bookingQuery).
		WithArgs(bookingID, userID, flightID, booking.Status, pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// commit transaction
	mockDb.ExpectCommit()

	createdBooking, err := repo.CreateBooking(context.Background(), booking)
	require.NoError(t, err)
	// verify the important fields
	assert.Equal(t, booking.ID, createdBooking.ID)
	assert.Equal(t, booking.User.FirstName, createdBooking.User.FirstName)
	assert.Equal(t, booking.User.LastName, createdBooking.User.LastName)
	assert.Equal(t, booking.User.Gender, createdBooking.User.Gender)
	assert.Equal(t, booking.Flight.LaunchpadID, createdBooking.Flight.LaunchpadID)
	assert.Equal(t, booking.Flight.Destination.ID, createdBooking.Flight.Destination.ID)
	assert.Equal(t, models.StatusConfirmed, createdBooking.Status)

	// verify time fields are set (without comparing exact values)
	assert.False(t, createdBooking.CreatedAt.IsZero())
	assert.False(t, createdBooking.User.Birthday.IsZero())
	assert.False(t, createdBooking.Flight.LaunchDate.IsZero())

	// verify all expectations were met
	err = mockDb.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetBookingsPaginated(t *testing.T) {
	t.Run("successful query without cursor", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		limit := 2
		bookings := createMockBookings(2)

		rows := createMockRows(bookings)

		expectedQuery := `
            SELECT 
                B.id, B.status, B.created_at,
                U.id, U.first_name, U.last_name, U.gender, U.birthday,
                F.id, F.launchpad_id, F.launch_date,
                D.id, D.name
            FROM bookings B
            JOIN users U ON U.id = B.user_id
            JOIN flights F ON F.id = B.flight_id
            JOIN destinations D ON D.id = F.destination_id
            ORDER BY B.created_at, B.id
            LIMIT $1`

		mockDb.ExpectQuery(formatQueryForRegex(expectedQuery)).
			WithArgs(limit).
			WillReturnRows(rows)

		result, cursor, err := repo.GetBookingsPaginated(context.Background(), "", limit)

		require.NoError(t, err)
		require.Len(t, result, 2)
		assert.NotEmpty(t, cursor)
		verifyBookings(t, bookings, result)
	})

	t.Run("successful query with cursor", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		limit := 2
		bookings := createMockBookings(2)
		cursorTime := time.Now()
		cursorID := uuid.New()
		cursor := encodeCursor(cursorTime, cursorID)

		rows := createMockRows(bookings)

		expectedQuery := `
            SELECT 
                B.id, B.status, B.created_at,
                U.id, U.first_name, U.last_name, U.gender, U.birthday,
                F.id, F.launchpad_id, F.launch_date,
                D.id, D.name
            FROM bookings B
            JOIN users U ON U.id = B.user_id
            JOIN flights F ON F.id = B.flight_id
            JOIN destinations D ON D.id = F.destination_id
            WHERE (B.created_at, B.id) > ($1, $2)
            ORDER BY B.created_at, B.id
            LIMIT $3`

		mockDb.ExpectQuery(formatQueryForRegex(expectedQuery)).
			WithArgs(pgxmock.AnyArg(), cursorID, limit).
			WillReturnRows(rows)

		result, nextCursor, err := repo.GetBookingsPaginated(context.Background(), cursor, limit)

		require.NoError(t, err)
		require.Len(t, result, 2)
		assert.NotEmpty(t, nextCursor)
		verifyBookings(t, bookings, result)
	})

	t.Run("empty result", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		limit := 2
		rows := pgxmock.NewRows([]string{
			"id", "status", "created_at",
			"user_id", "first_name", "last_name", "gender", "birthday",
			"flight_id", "launchpad_id", "launch_date",
			"destination_id", "destination_name",
		})
		expectedQuery := `
			SELECT 
				B.id, B.status, B.created_at,
				U.id, U.first_name, U.last_name, U.gender, U.birthday,
				F.id, F.launchpad_id, F.launch_date,
				D.id, D.name
			FROM bookings B
			JOIN users U ON U.id = B.user_id
			JOIN flights F ON F.id = B.flight_id
			JOIN destinations D ON D.id = F.destination_id
			ORDER BY B.created_at, B.id
			LIMIT $1`

		mockDb.ExpectQuery(formatQueryForRegex(expectedQuery)).
			WithArgs(limit).
			WillReturnRows(rows)

		result, cursor, err := repo.GetBookingsPaginated(context.Background(), "", limit)

		require.NoError(t, err)
		assert.Empty(t, result)
		assert.Empty(t, cursor)
	})

	t.Run("invalid cursor format", func(t *testing.T) {
		_, repo := setupMockDB(t)

		invalidCursor := base64.StdEncoding.EncodeToString([]byte("invalid"))

		_, _, err := repo.GetBookingsPaginated(context.Background(), invalidCursor, 10)
		assert.Error(t, err)
	})

	t.Run("database error", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		mockDb.ExpectQuery(formatQueryForRegex(`SELECT.*FROM bookings.*`)).
			WithArgs(10).
			WillReturnError(fmt.Errorf("database error"))

		_, _, err := repo.GetBookingsPaginated(context.Background(), "", 10)
		assert.Error(t, err)
	})

	t.Run("scan error", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		rows := pgxmock.NewRows([]string{"id"}).AddRow("invalid") // This will cause a scan error

		mockDb.ExpectQuery(formatQueryForRegex(`SELECT.*FROM bookings.*`)).
			WithArgs(10).
			WillReturnRows(rows)

		_, _, err := repo.GetBookingsPaginated(context.Background(), "", 10)
		assert.Error(t, err)
	})
}

func TestGetBookingByID(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		expectedBooking := createMockBookings(1)[0]
		rows := createMockRows([]models.Booking{expectedBooking})

		mockDb.ExpectQuery("SELECT.*FROM bookings.*WHERE B.id = \\$1").
			WithArgs(expectedBooking.ID.String()).
			WillReturnRows(rows)

		booking, err := repo.GetBookingByID(context.Background(), expectedBooking.ID.String())

		assert.NoError(t, err)
		assert.Equal(t, expectedBooking.ID, booking.ID)
		assert.NoError(t, mockDb.ExpectationsWereMet())
	})

	t.Run("booking not found", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		bookingID := uuid.New()

		mockDb.ExpectQuery("SELECT.*FROM bookings.*WHERE B.id = \\$1").
			WithArgs(bookingID.String()).
			WillReturnError(pgx.ErrNoRows)

		booking, err := repo.GetBookingByID(context.Background(), bookingID.String())

		assert.Error(t, err)
		assert.Equal(t, models.ErrBookingNotFound, err)
		assert.Nil(t, booking)
		assert.NoError(t, mockDb.ExpectationsWereMet())
	})
}

func TestGetDestinationById(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		destID := uuid.New()
		expectedDest := models.Destination{
			ID:   destID,
			Name: "Mars",
		}

		mockDb.ExpectQuery("SELECT id, name FROM destinations WHERE id = \\$1").
			WithArgs(destID.String()).
			WillReturnRows(pgxmock.NewRows([]string{"id", "name"}).
				AddRow(expectedDest.ID, expectedDest.Name))

		result, err := repo.GetDestinationById(context.Background(), destID.String())

		require.NoError(t, err)
		assert.Equal(t, expectedDest.ID, result.ID)
		assert.Equal(t, expectedDest.Name, result.Name)

		err = mockDb.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("destination not found", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		nonExistentID := uuid.New()

		mockDb.ExpectQuery("SELECT id, name FROM destinations WHERE id = \\$1").
			WithArgs(nonExistentID.String()).
			WillReturnError(pgx.ErrNoRows)

		result, err := repo.GetDestinationById(context.Background(), nonExistentID.String())

		assert.Error(t, err)
		assert.Equal(t, pgx.ErrNoRows, err)
		assert.Empty(t, result)

		err = mockDb.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("invalid uuid format", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		invalidID := "not-a-uuid"

		mockDb.ExpectQuery("SELECT id, name FROM destinations WHERE id = \\$1").
			WithArgs(invalidID).
			WillReturnError(errors.New("invalid UUID format"))

		result, err := repo.GetDestinationById(context.Background(), invalidID)

		assert.Error(t, err)
		assert.Empty(t, result)

		err = mockDb.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("database error", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		destID := uuid.New()

		mockDb.ExpectQuery("SELECT id, name FROM destinations WHERE id = \\$1").
			WithArgs(destID.String()).
			WillReturnError(errors.New("database connection error"))

		result, err := repo.GetDestinationById(context.Background(), destID.String())

		assert.Error(t, err)
		assert.Empty(t, result)

		err = mockDb.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("context canceled", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		destID := uuid.New()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		mockDb.ExpectQuery("SELECT id, name FROM destinations WHERE id = \\$1").
			WithArgs(destID.String()).
			WillReturnError(context.Canceled)

		result, err := repo.GetDestinationById(ctx, destID.String())

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
		assert.Empty(t, result)

		err = mockDb.ExpectationsWereMet()
		require.NoError(t, err)
	})
}

func TestDeleteBooking(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		bookingID := uuid.New().String()

		mockDb.ExpectBegin()
		mockDb.ExpectExec("DELETE FROM bookings WHERE id = \\$1").
			WithArgs(bookingID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))
		mockDb.ExpectCommit()

		err := repo.DeleteBooking(context.Background(), bookingID)

		assert.NoError(t, err)
		assert.NoError(t, mockDb.ExpectationsWereMet())
	})

	t.Run("booking not found", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		bookingID := uuid.New().String()

		mockDb.ExpectBegin()
		mockDb.ExpectExec("DELETE FROM bookings WHERE id = \\$1").
			WithArgs(bookingID).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))
		mockDb.ExpectRollback()

		err := repo.DeleteBooking(context.Background(), bookingID)

		assert.Error(t, err)
		assert.Equal(t, models.ErrBookingNotFound, err)
		assert.NoError(t, mockDb.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		bookingID := uuid.New().String()

		mockDb.ExpectBegin()
		mockDb.ExpectExec("DELETE FROM bookings WHERE id = \\$1").
			WithArgs(bookingID).
			WillReturnError(errors.New("database error"))
		mockDb.ExpectRollback()

		err := repo.DeleteBooking(context.Background(), bookingID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete booking")
		assert.NoError(t, mockDb.ExpectationsWereMet())
	})
}

func TestBookingRepository_GetFlights(t *testing.T) {
	t.Run("successful retrieval without filters", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		expectedFlights := []models.Flight{
			{
				ID:          uuid.New(),
				LaunchpadID: "LP1",
				LaunchDate:  time.Now().Add(24 * time.Hour),
				Destination: models.Destination{
					ID:   uuid.New(),
					Name: "Mars",
				},
			},
			{
				ID:          uuid.New(),
				LaunchpadID: "LP2",
				LaunchDate:  time.Now().Add(48 * time.Hour),
				Destination: models.Destination{
					ID:   uuid.New(),
					Name: "Moon",
				},
			},
		}

		mockDb.ExpectBegin()

		rows := pgxmock.NewRows([]string{
			"id", "launchpad_id", "launch_date",
			"destination_id", "destination_name",
		})
		for _, f := range expectedFlights {
			rows.AddRow(
				f.ID, f.LaunchpadID, f.LaunchDate,
				f.Destination.ID, f.Destination.Name,
			)
		}

		mockDb.ExpectQuery(`SELECT F.id, F.launchpad_id, F.launch_date,
            D.id as destination_id, D.name as destination_name
            FROM flights F
            JOIN destinations D ON D.id = F.destination_id`).
			WillReturnRows(rows)

		mockDb.ExpectCommit()

		flights, err := repo.GetFlights(context.Background(), map[string]interface{}{})

		require.NoError(t, err)
		assert.Len(t, flights, 2)
		assert.Equal(t, expectedFlights[0].LaunchpadID, flights[0].LaunchpadID)
		assert.Equal(t, expectedFlights[1].LaunchpadID, flights[1].LaunchpadID)
	})

	t.Run("with filters", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		filters := map[string]interface{}{
			"launchpad_id":    "LP1",
			"bookings.status": "CONFIRMED",
		}

		expectedFlight := models.Flight{
			ID:          uuid.New(),
			LaunchpadID: "LP1",
			LaunchDate:  time.Now().Add(24 * time.Hour),
			Destination: models.Destination{
				ID:   uuid.New(),
				Name: "Mars",
			},
		}

		mockDb.ExpectBegin()

		rows := pgxmock.NewRows([]string{
			"id", "launchpad_id", "launch_date",
			"destination_id", "destination_name",
		}).AddRow(
			expectedFlight.ID, expectedFlight.LaunchpadID, expectedFlight.LaunchDate,
			expectedFlight.Destination.ID, expectedFlight.Destination.Name,
		)

		mockDb.ExpectQuery(`SELECT F.id, F.launchpad_id, F.launch_date,
            D.id as destination_id, D.name as destination_name
            FROM flights F
            JOIN destinations D ON D.id = F.destination_id
            JOIN bookings B ON B.flight_id = F.id
            WHERE F.launchpad_id=\$1 AND B.status=\$2
            GROUP BY F.id, D.id`).
			WithArgs("LP1", "CONFIRMED").
			WillReturnRows(rows)

		mockDb.ExpectCommit()

		flights, err := repo.GetFlights(context.Background(), filters)

		require.NoError(t, err)
		assert.Len(t, flights, 1)
		assert.Equal(t, expectedFlight.LaunchpadID, flights[0].LaunchpadID)
	})

	t.Run("transaction begin error", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		mockDb.ExpectBegin().WillReturnError(errors.New("begin error"))

		flights, err := repo.GetFlights(context.Background(), map[string]interface{}{})

		assert.Error(t, err)
		assert.Nil(t, flights)
	})

	t.Run("query error", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		mockDb.ExpectBegin()
		mockDb.ExpectQuery(`SELECT F.id, F.launchpad_id`).
			WillReturnError(errors.New("query error"))
		mockDb.ExpectRollback()

		flights, err := repo.GetFlights(context.Background(), map[string]interface{}{})

		assert.Error(t, err)
		assert.Nil(t, flights)
	})

	t.Run("scan error", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		mockDb.ExpectBegin()

		rows := pgxmock.NewRows([]string{
			"id", "launchpad_id", "launch_date",
			"destination_id", "destination_name",
		}).AddRow(
			"invalid-uuid", "LP1", time.Now(),
			uuid.New(), "Mars",
		)

		mockDb.ExpectQuery(`SELECT F.id, F.launchpad_id`).
			WillReturnRows(rows)
		mockDb.ExpectRollback()

		flights, err := repo.GetFlights(context.Background(), map[string]interface{}{})

		assert.Error(t, err)
		assert.Nil(t, flights)
	})

	t.Run("commit error", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		mockDb.ExpectBegin()
		mockDb.ExpectQuery(`SELECT F.id, F.launchpad_id`).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "launchpad_id", "launch_date",
				"destination_id", "destination_name",
			}))
		mockDb.ExpectCommit().WillReturnError(errors.New("commit error"))

		flights, err := repo.GetFlights(context.Background(), map[string]interface{}{})

		assert.Error(t, err)
		assert.Nil(t, flights)
	})

	t.Run("context cancellation", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		mockDb.ExpectBegin().WillReturnError(context.Canceled)

		flights, err := repo.GetFlights(ctx, map[string]interface{}{})

		assert.Equal(t, context.Canceled, err)
		assert.Nil(t, flights)
	})
}

func TestBookingRepository_IsLaunchPadWeekAvailable(t *testing.T) {
	t.Run("launchpad is available", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		launchpadID := "LP1"
		destinationID := uuid.New().String()
		launchTime := time.Now()

		mockDb.ExpectBegin()

		rows := pgxmock.NewRows([]string{"launch_in_same_week"}).
			AddRow(true)

		mockDb.ExpectQuery("SELECT launch_in_same_week\\(\\$1, \\$2, \\$3\\)").
			WithArgs(launchpadID, destinationID, launchTime).
			WillReturnRows(rows)

		mockDb.ExpectCommit()

		available, err := repo.IsLaunchPadWeekAvailable(
			context.Background(),
			launchpadID,
			destinationID,
			launchTime,
		)

		require.NoError(t, err)
		assert.True(t, available)

		err = mockDb.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("launchpad is not available", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		launchpadID := "LP1"
		destinationID := uuid.New().String()
		launchTime := time.Now()

		mockDb.ExpectBegin()

		rows := pgxmock.NewRows([]string{"launch_in_same_week"}).
			AddRow(false)

		mockDb.ExpectQuery("SELECT launch_in_same_week\\(\\$1, \\$2, \\$3\\)").
			WithArgs(launchpadID, destinationID, launchTime).
			WillReturnRows(rows)

		mockDb.ExpectCommit()

		available, err := repo.IsLaunchPadWeekAvailable(
			context.Background(),
			launchpadID,
			destinationID,
			launchTime,
		)

		require.NoError(t, err)
		assert.False(t, available)

		err = mockDb.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("transaction begin error", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		launchpadID := "LP1"
		destinationID := uuid.New().String()
		launchTime := time.Now()

		mockDb.ExpectBegin().WillReturnError(errors.New("begin error"))

		available, err := repo.IsLaunchPadWeekAvailable(
			context.Background(),
			launchpadID,
			destinationID,
			launchTime,
		)

		assert.Error(t, err)
		assert.False(t, available)
		assert.Equal(t, "begin error", err.Error())
	})

	t.Run("query error", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		launchpadID := "LP1"
		destinationID := uuid.New().String()
		launchTime := time.Now()

		mockDb.ExpectBegin()

		mockDb.ExpectQuery("SELECT launch_in_same_week\\(\\$1, \\$2, \\$3\\)").
			WithArgs(launchpadID, destinationID, launchTime).
			WillReturnError(errors.New("query error"))

		mockDb.ExpectRollback()

		available, err := repo.IsLaunchPadWeekAvailable(
			context.Background(),
			launchpadID,
			destinationID,
			launchTime,
		)

		assert.Error(t, err)
		assert.False(t, available)
		assert.Equal(t, "query error", err.Error())
	})

	t.Run("scan error", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		launchpadID := "LP1"
		destinationID := uuid.New().String()
		launchTime := time.Now()

		mockDb.ExpectBegin()

		// return invalid type that will cause scan error
		rows := pgxmock.NewRows([]string{"launch_in_same_week"}).
			AddRow("not a boolean")

		mockDb.ExpectQuery("SELECT launch_in_same_week\\(\\$1, \\$2, \\$3\\)").
			WithArgs(launchpadID, destinationID, launchTime).
			WillReturnRows(rows)

		mockDb.ExpectRollback()

		available, err := repo.IsLaunchPadWeekAvailable(
			context.Background(),
			launchpadID,
			destinationID,
			launchTime,
		)

		assert.Error(t, err)
		assert.False(t, available)
	})

	t.Run("commit error", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		launchpadID := "LP1"
		destinationID := uuid.New().String()
		launchTime := time.Now()

		mockDb.ExpectBegin()

		rows := pgxmock.NewRows([]string{"launch_in_same_week"}).
			AddRow(true)

		mockDb.ExpectQuery("SELECT launch_in_same_week\\(\\$1, \\$2, \\$3\\)").
			WithArgs(launchpadID, destinationID, launchTime).
			WillReturnRows(rows)

		mockDb.ExpectCommit().WillReturnError(errors.New("commit error"))

		available, err := repo.IsLaunchPadWeekAvailable(
			context.Background(),
			launchpadID,
			destinationID,
			launchTime,
		)

		assert.Error(t, err)
		assert.True(t, available) // The query returns true, even though commit fails
		assert.Equal(t, "commit error", err.Error())

		err = mockDb.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("rollback after query error", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		launchpadID := "LP1"
		destinationID := uuid.New().String()
		launchTime := time.Now()

		mockDb.ExpectBegin()

		mockDb.ExpectQuery("SELECT launch_in_same_week\\(\\$1, \\$2, \\$3\\)").
			WithArgs(launchpadID, destinationID, launchTime).
			WillReturnError(errors.New("query error"))

		mockDb.ExpectRollback()

		available, err := repo.IsLaunchPadWeekAvailable(
			context.Background(),
			launchpadID,
			destinationID,
			launchTime,
		)

		assert.Error(t, err)
		assert.False(t, available) // query error results in false
		assert.Equal(t, "query error", err.Error())

		err = mockDb.ExpectationsWereMet()
		require.NoError(t, err)
	})

	t.Run("context canceled", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		launchpadID := "LP1"
		destinationID := uuid.New().String()
		launchTime := time.Now()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		mockDb.ExpectBegin().WillReturnError(context.Canceled)

		available, err := repo.IsLaunchPadWeekAvailable(
			ctx,
			launchpadID,
			destinationID,
			launchTime,
		)

		assert.Error(t, err)
		assert.False(t, available)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("with nil values", func(t *testing.T) {
		mockDb, repo := setupMockDB(t)
		defer mockDb.Close()

		mockDb.ExpectBegin()

		rows := pgxmock.NewRows([]string{"launch_in_same_week"}).
			AddRow(nil)

		mockDb.ExpectQuery("SELECT launch_in_same_week\\(\\$1, \\$2, \\$3\\)").
			WithArgs(nil, nil, time.Time{}).
			WillReturnRows(rows)

		mockDb.ExpectRollback()

		available, err := repo.IsLaunchPadWeekAvailable(
			context.Background(),
			"",
			"",
			time.Time{},
		)

		assert.Error(t, err)
		assert.False(t, available)
	})
}

// helper functions
func setupMockDB(t *testing.T) (pgxmock.PgxPoolIface, *repository.BookingRepository) {
	mockDb, err := pgxmock.NewPool()
	require.NoError(t, err)
	return mockDb, repository.NewBookingRepository(mockDb)
}

func createMockBookings(count int) []models.Booking {
	bookings := make([]models.Booking, count)
	for i := 0; i < count; i++ {
		bookings[i] = models.Booking{
			ID:        uuid.New(),
			Status:    models.StatusConfirmed,
			CreatedAt: time.Now().Add(time.Duration(i) * time.Hour),
			User: models.User{
				ID:        uuid.New(),
				FirstName: fmt.Sprintf("User%d", i),
				LastName:  "Doe",
				Gender:    "Male",
				Birthday:  time.Now().Add(-20 * 365 * 24 * time.Hour),
			},
			Flight: models.Flight{
				ID:          uuid.New(),
				LaunchpadID: fmt.Sprintf("LP%d", i),
				LaunchDate:  time.Now().Add(24 * time.Hour),
				Destination: models.Destination{
					ID:   uuid.New(),
					Name: fmt.Sprintf("Destination%d", i),
				},
			},
		}
	}
	return bookings
}

func createMockRows(bookings []models.Booking) *pgxmock.Rows {
	rows := pgxmock.NewRows([]string{
		"id", "status", "created_at",
		"user_id", "first_name", "last_name", "gender", "birthday",
		"flight_id", "launchpad_id", "launch_date",
		"destination_id", "destination_name",
	})

	for _, b := range bookings {
		rows.AddRow(
			b.ID, b.Status, b.CreatedAt,
			b.User.ID, b.User.FirstName, b.User.LastName, b.User.Gender, b.User.Birthday,
			b.Flight.ID, b.Flight.LaunchpadID, b.Flight.LaunchDate,
			b.Flight.Destination.ID, b.Flight.Destination.Name,
		)
	}
	return rows
}

func verifyBookings(t *testing.T, expected, actual []models.Booking) {
	require.Equal(t, len(expected), len(actual))
	for i := range expected {
		assert.Equal(t, expected[i].ID, actual[i].ID)
		assert.Equal(t, expected[i].Status, actual[i].Status)
		assert.Equal(t, expected[i].User.FirstName, actual[i].User.FirstName)
		assert.Equal(t, expected[i].User.LastName, actual[i].User.LastName)
		assert.Equal(t, expected[i].Flight.LaunchpadID, actual[i].Flight.LaunchpadID)
		assert.Equal(t, expected[i].Flight.Destination.Name, actual[i].Flight.Destination.Name)
	}
}

func formatQueryForRegex(query string) string {
	// remove extra whitespace and newlines
	query = strings.Join(strings.Fields(query), " ")
	// escape special regex characters
	query = regexp.QuoteMeta(query)
	return fmt.Sprintf("^%s$", query)
}

func encodeCursor(t time.Time, id uuid.UUID) string {
	cursor := fmt.Sprintf("%s,%s", t.Format(time.RFC3339Nano), id.String())
	return base64.StdEncoding.EncodeToString([]byte(cursor))
}
