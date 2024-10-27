package service

import (
	"context"
	"fmt"
	models "github.com/chrisdamba/spacetrouble/internal"
	"github.com/chrisdamba/spacetrouble/internal/ports"
	"github.com/google/uuid"
	"time"
)

type bookingService struct {
	repo   ports.BookingRepository
	spaceX ports.SpaceXClient
}

func NewBookingService(repo ports.BookingRepository, spaceX ports.SpaceXClient) *bookingService {
	return &bookingService{
		repo:   repo,
		spaceX: spaceX,
	}
}

func (s *bookingService) CreateBooking(ctx context.Context, request *models.BookingRequest) (*models.Booking, error) {
	// validate the destination exists
	destination, err := s.repo.GetDestinationById(ctx, request.DestinationID.String())
	if err != nil {
		return nil, fmt.Errorf("invalid destination: %w", err)
	}

	// check if launchpad is already booked for this date
	flights, err := s.repo.GetFlights(ctx, map[string]interface{}{
		"launchpad_id": request.LaunchpadID,
		"launch_date":  request.LaunchDate,
	})
	if err != nil {
		return nil, fmt.Errorf("error checking launchpad availability: %w", err)
	}

	// if flights exist for this date but different destination, launchpad is unavailable
	if len(flights) > 0 && flights[0].Destination.ID != request.DestinationID {
		return nil, fmt.Errorf("launchpad already booked for different destination on this date")
	}

	// check if launchpad is already used for this destination in the same week
	available, err := s.repo.IsLaunchPadWeekAvailable(ctx,
		request.LaunchpadID,
		request.DestinationID.String(),
		request.LaunchDate)
	if err != nil {
		return nil, fmt.Errorf("error checking weekly availability: %w", err)
	}
	if !available {
		return nil, fmt.Errorf("launchpad already scheduled for this destination this week")
	}

	// check SpaceX launch conflict
	spaceXAvailable, err := s.spaceX.CheckLaunchConflict(ctx, request.LaunchpadID, request.LaunchDate)
	if err != nil {
		return nil, fmt.Errorf("error checking SpaceX availability: %w", err)
	}
	if !spaceXAvailable {
		return nil, fmt.Errorf("launchpad reserved by SpaceX on this date")
	}

	// create the booking
	booking := &models.Booking{
		ID: uuid.New(),
		User: models.User{
			ID:        uuid.New(),
			FirstName: request.FirstName,
			LastName:  request.LastName,
			Gender:    request.Gender,
			Birthday:  request.Birthday,
		},
		Flight: models.Flight{
			ID:          uuid.New(),
			LaunchpadID: request.LaunchpadID,
			Destination: *destination,
			LaunchDate:  request.LaunchDate,
		},
		Status:    models.StatusActive,
		CreatedAt: time.Now().UTC(),
	}

	// persist to db
	savedBooking, err := s.repo.CreateBooking(ctx, booking)
	if err != nil {
		return nil, fmt.Errorf("error creating booking: %w", err)
	}

	return savedBooking, nil
}

func (s *bookingService) AllBookings(ctx context.Context, req models.GetBookingsRequest) (*models.AllBookingsResponse, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	bookings, nextCursor, err := s.repo.GetBookingsPaginated(ctx, req.Uuid, limit)
	if err != nil {
		return nil, fmt.Errorf("error fetching bookings: %w", err)
	}

	response := &models.AllBookingsResponse{
		Bookings: make([]models.BookingResponse, len(bookings)),
		Limit:    limit,
		Cursor:   nextCursor,
	}

	for i, booking := range bookings {
		response.Bookings[i] = models.BookingResponse{Booking: booking}
	}

	return response, nil
}

func (s *bookingService) DeleteBooking(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return models.ErrInvalidUUID
	}

	booking, err := s.repo.GetBookingByID(ctx, id)
	if err != nil {
		return err
	}

	// only allow deletion of active or confirmed bookings
	if booking.Status != models.StatusActive && booking.Status != models.StatusConfirmed {
		return fmt.Errorf("cannot delete booking with status %s", booking.Status)
	}

	return s.repo.DeleteBooking(ctx, id)
}
