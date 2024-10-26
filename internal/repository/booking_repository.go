package repository

import (
	"context"
	"encoding/base64"
	"fmt"
	models "github.com/chrisdamba/spacetrouble/internal"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"strings"
	"time"
)

type DBConn interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
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

func (r *BookingRepository) GetBookingsPaginated(ctx context.Context, afterCursor string, limit int) ([]models.Booking, string, error) {
	query := `
        SELECT 
            B.id, B.status, B.created_at,
            U.id, U.first_name, U.last_name, U.gender, U.birthday,
            F.id, F.launchpad_id, F.launch_date,
            D.id, D.name
        FROM bookings B
        JOIN users U ON U.id = B.user_id
        JOIN flights F ON F.id = B.flight_id
        JOIN destinations D ON D.id = F.destination_id
    `
	var args []interface{}
	var conditions []string

	if afterCursor != "" {
		afterTime, afterUUID, err := decodeCursor(afterCursor)
		if err != nil {
			return nil, "", err
		}
		conditions = append(conditions, "(B.created_at, B.id) > ($1, $2)")
		args = append(args, afterTime, afterUUID)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY B.created_at, B.id"
	query += fmt.Sprintf(" LIMIT $%d", len(args)+1)
	args = append(args, limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var bookings []models.Booking
	var lastBooking models.Booking
	for rows.Next() {
		var booking models.Booking
		var destinationID uuid.UUID
		var destinationName string

		err := rows.Scan(
			&booking.ID, &booking.Status, &booking.CreatedAt,
			&booking.User.ID, &booking.User.FirstName, &booking.User.LastName, &booking.User.Gender, &booking.User.Birthday,
			&booking.Flight.ID, &booking.Flight.LaunchpadID, &booking.Flight.LaunchDate,
			&destinationID, &destinationName,
		)
		if err != nil {
			return nil, "", err
		}
		booking.Flight.Destination = models.Destination{
			ID:   destinationID,
			Name: destinationName,
		}
		bookings = append(bookings, booking)
		lastBooking = booking
	}
	if err = rows.Err(); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(bookings) == limit {
		nextCursor = encodeCursor(lastBooking.CreatedAt, lastBooking.ID)
	}

	return bookings, nextCursor, nil
}

func (r *BookingRepository) GetDestinationById(ctx context.Context, id string) (*models.Destination, error) {
	q := `SELECT id, name FROM destinations WHERE id = $1`
	var dest models.Destination
	if err := r.db.QueryRow(ctx, q, id).Scan(&dest.ID, &dest.Name); err != nil {
		return &dest, err
	}
	return &dest, nil

}

func (r *BookingRepository) GetFlights(ctx context.Context, filters map[string]interface{}) ([]models.Flight, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	flights, err := r.selectFlightsTx(ctx, tx, filters)
	if err != nil {
		return nil, err
	}
	return flights, tx.Commit(ctx)
}

func (r *BookingRepository) IsLaunchPadWeekAvailable(ctx context.Context, launchpadId, destinationId string,
	t time.Time) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	ans, err := r.getLaunchPadWeekAvailabiltyTx(ctx, tx, launchpadId, destinationId, t)
	if err != nil {
		return ans, err
	}
	return ans, tx.Commit(ctx)
}

func (r *BookingRepository) getLaunchPadWeekAvailabiltyTx(ctx context.Context, tx pgx.Tx,
	launchpadId, destinationId string, t time.Time) (bool, error) {
	var ans bool
	err := tx.QueryRow(ctx, `SELECT launch_in_same_week($1, $2, $3)`, launchpadId, destinationId, t).Scan(&ans)

	return ans, err
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

func (r *BookingRepository) buildSelectFlightQuery(filters map[string]interface{}) (string, []interface{}) {
	q := `SELECT 
			F.id, F.launchpad_id, F.launch_date,
			D.id as destination_id, D.name as destination_name
			FROM flights F
			JOIN destinations D ON D.id = F.destination_id`
	bookingStatus, hasBookingStatus := filters["bookings.status"]
	if hasBookingStatus {
		q += ` JOIN bookings B ON B.flight_id = F.id`
	}
	var whereConds []string
	var args []interface{}
	for k, v := range filters {
		if !strings.HasPrefix(k, "bookings.") {
			whereConds = append(whereConds, fmt.Sprintf("F.%s=$%d", k, len(args)+1))
			args = append(args, v)
		}
	}
	if hasBookingStatus {
		whereConds = append(whereConds, fmt.Sprintf("B.status=$%d", len(args)+1))
		args = append(args, bookingStatus)
	}
	if len(whereConds) > 0 {
		q += " WHERE " + strings.Join(whereConds, " AND ")
	}
	if hasBookingStatus {
		q += " GROUP BY F.id, D.id"
	}
	return q, args
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

func (r *BookingRepository) selectFlightsTx(ctx context.Context, tx pgx.Tx, filters map[string]interface{}) ([]models.Flight, error) {
	q, args := r.buildSelectFlightQuery(filters)
	rows, err := tx.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.Flight
	for rows.Next() {
		var flight models.Flight
		err := rows.Scan(
			&flight.ID, &flight.LaunchpadID, &flight.LaunchDate,
			&flight.Destination.ID, &flight.Destination.Name,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, flight)
	}
	return items, rows.Err()
}

func (r *BookingRepository) createBookingTx(ctx context.Context, tx pgx.Tx, booking *models.Booking) error {
	query := `
        INSERT INTO bookings (id, user_id, flight_id, status, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := tx.Exec(ctx, query, booking.ID, booking.User.ID, booking.Flight.ID, booking.Status, booking.CreatedAt)
	return err
}

func encodeCursor(t time.Time, id uuid.UUID) string {
	cursor := fmt.Sprintf("%s,%s", t.Format(time.RFC3339Nano), id.String())
	return base64.StdEncoding.EncodeToString([]byte(cursor))
}

func decodeCursor(encoded string) (time.Time, uuid.UUID, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	parts := strings.Split(string(decodedBytes), ",")
	if len(parts) != 2 {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor format")
	}
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	return t, id, nil
}
