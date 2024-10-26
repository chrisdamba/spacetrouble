package repository

import (
	"context"
	models "github.com/chrisdamba/spacetrouble/internal"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"time"
)

type DBConn interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

type BookingRepository struct {
	db DBConn
}

func NewBookingRepository(db DBConn) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) CreateBooking(ctx context.Context, booking *models.Booking) (*models.Booking, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// create User if not exists
	err = r.createUserTx(ctx, tx, &booking.User)
	if err != nil {
		return nil, err
	}

	// create Flight if not exists
	err = r.createFlightTx(ctx, tx, &booking.Flight)
	if err != nil {
		return nil, err
	}

	// create Booking
	if booking.ID == uuid.Nil {
		booking.ID = uuid.New()
	}
	booking.Status = models.StatusConfirmed
	booking.CreatedAt = time.Now().UTC()
	err = r.createBookingTx(ctx, tx, booking)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}
	return booking, nil
}

func (r *BookingRepository) createUserTx(ctx context.Context, tx pgx.Tx, user *models.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	query := `
        INSERT INTO users (id, first_name, last_name, gender, birthday)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (id) DO NOTHING
    `
	_, err := tx.Exec(ctx, query, user.ID, user.FirstName, user.LastName, user.Gender, user.Birthday)
	return err
}

func (r *BookingRepository) createFlightTx(ctx context.Context, tx pgx.Tx, flight *models.Flight) error {
	if flight.ID == uuid.Nil {
		flight.ID = uuid.New()
	}
	query := `
        INSERT INTO flights (id, launchpad_id, destination_id, launch_date)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id) DO NOTHING
    `
	_, err := tx.Exec(ctx, query, flight.ID, flight.LaunchpadID, flight.Destination.ID, flight.LaunchDate)
	return err
}

func (r *BookingRepository) createBookingTx(ctx context.Context, tx pgx.Tx, booking *models.Booking) error {
	query := `
        INSERT INTO bookings (id, user_id, flight_id, status, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := tx.Exec(ctx, query, booking.ID, booking.User.ID, booking.Flight.ID, booking.Status, booking.CreatedAt)
	return err
}
