package service_test

import (
	"context"
	"errors"
	models "github.com/chrisdamba/spacetrouble/internal"
	"github.com/chrisdamba/spacetrouble/internal/service"
	"github.com/chrisdamba/spacetrouble/tests/mocks"
	"github.com/chrisdamba/spacetrouble/tests/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestCreateBooking(t *testing.T) {
	validDestinationID := uuid.New()
	validLaunchDate := time.Now().Add(24 * time.Hour)

	validRequest := &models.BookingRequest{
		FirstName:     "John",
		LastName:      "Doe",
		Gender:        "Male",
		Birthday:      time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		LaunchpadID:   "pad-1",
		DestinationID: validDestinationID,
		LaunchDate:    validLaunchDate,
	}

	validDestination := &models.Destination{
		ID:   validDestinationID,
		Name: "Mars",
	}

	t.Run("Successful booking creation", func(t *testing.T) {
		mockRepo := new(mocks.MockBookingRepository)
		mockSpaceX := new(mocks.MockSpaceXClient)
		svc := service.NewBookingService(mockRepo, mockSpaceX)
		ctx := context.Background()

		mockRepo.On("GetDestinationById", ctx, validDestinationID.String()).Return(validDestination, nil)
		mockRepo.On("GetFlights", ctx, mock.Anything).Return([]models.Flight{}, nil)
		mockRepo.On("IsLaunchPadWeekAvailable", ctx, "pad-1", validDestinationID.String(), validLaunchDate).Return(true, nil)
		mockSpaceX.On("CheckLaunchConflict", ctx, "pad-1", validLaunchDate).Return(true, nil)

		mockRepo.On("CreateBooking", ctx, mock.AnythingOfType("*models.Booking")).
			Run(func(args mock.Arguments) {
			}).
			Return(&models.Booking{
				ID: uuid.New(),
				User: models.User{
					ID:        uuid.New(),
					FirstName: validRequest.FirstName,
					LastName:  validRequest.LastName,
					Gender:    validRequest.Gender,
					Birthday:  validRequest.Birthday,
				},
				Flight: models.Flight{
					ID:          uuid.New(),
					LaunchpadID: validRequest.LaunchpadID,
					Destination: *validDestination,
					LaunchDate:  validRequest.LaunchDate,
				},
				Status:    models.StatusActive,
				CreatedAt: time.Now().UTC(),
			}, nil)

		booking, err := svc.CreateBooking(ctx, validRequest)

		assert.NoError(t, err)
		assert.NotNil(t, booking)
		assert.Equal(t, validRequest.FirstName, booking.User.FirstName)
		assert.Equal(t, validRequest.LastName, booking.User.LastName)
		assert.Equal(t, validRequest.LaunchpadID, booking.Flight.LaunchpadID)
		assert.Equal(t, validDestinationID, booking.Flight.Destination.ID)
		assert.Equal(t, models.StatusActive, booking.Status)
		mockRepo.AssertExpectations(t)
		mockSpaceX.AssertExpectations(t)
	})

	t.Run("Invalid destination", func(t *testing.T) {
		mockRepo := new(mocks.MockBookingRepository)
		mockSpaceX := new(mocks.MockSpaceXClient)
		svc := service.NewBookingService(mockRepo, mockSpaceX)
		ctx := context.Background()

		mockRepo.On("GetDestinationById", ctx, validDestinationID.String()).Return(nil, assert.AnError)

		booking, err := svc.CreateBooking(ctx, validRequest)

		assert.Error(t, err)
		assert.Nil(t, booking)
		assert.Contains(t, err.Error(), "invalid destination")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Launchpad already booked for different destination", func(t *testing.T) {
		mockRepo := new(mocks.MockBookingRepository)
		mockSpaceX := new(mocks.MockSpaceXClient)
		svc := service.NewBookingService(mockRepo, mockSpaceX)
		ctx := context.Background()

		differentDestID := uuid.New()
		existingFlight := []models.Flight{{
			ID:          uuid.New(),
			LaunchpadID: "pad-1",
			Destination: models.Destination{ID: differentDestID},
			LaunchDate:  validLaunchDate,
		}}

		mockRepo.On("GetDestinationById", ctx, validDestinationID.String()).Return(validDestination, nil)
		mockRepo.On("GetFlights", ctx, mock.Anything).Return(existingFlight, nil)

		booking, err := svc.CreateBooking(ctx, validRequest)

		assert.Error(t, err)
		assert.Nil(t, booking)
		assert.Contains(t, err.Error(), "launchpad already booked")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Weekly launchpad unavailable", func(t *testing.T) {
		mockRepo := new(mocks.MockBookingRepository)
		mockSpaceX := new(mocks.MockSpaceXClient)
		svc := service.NewBookingService(mockRepo, mockSpaceX)
		ctx := context.Background()

		mockRepo.On("GetDestinationById", ctx, validDestinationID.String()).Return(validDestination, nil)
		mockRepo.On("GetFlights", ctx, mock.Anything).Return([]models.Flight{}, nil)
		mockRepo.On("IsLaunchPadWeekAvailable", ctx, "pad-1", validDestinationID.String(), validLaunchDate).Return(false, nil)

		booking, err := svc.CreateBooking(ctx, validRequest)

		assert.Error(t, err)
		assert.Nil(t, booking)
		assert.Contains(t, err.Error(), "already scheduled for this destination this week")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Database error during creation", func(t *testing.T) {
		mockRepo := new(mocks.MockBookingRepository)
		mockSpaceX := new(mocks.MockSpaceXClient)
		svc := service.NewBookingService(mockRepo, mockSpaceX)
		ctx := context.Background()

		mockRepo.On("GetDestinationById", ctx, validDestinationID.String()).Return(validDestination, nil)
		mockRepo.On("GetFlights", ctx, mock.Anything).Return([]models.Flight{}, nil)
		mockRepo.On("IsLaunchPadWeekAvailable", ctx, "pad-1", validDestinationID.String(), validLaunchDate).Return(true, nil)
		mockSpaceX.On("CheckLaunchConflict", ctx, "pad-1", validLaunchDate).Return(true, nil)
		mockRepo.On("CreateBooking", ctx, mock.AnythingOfType("*models.Booking")).Return(nil, assert.AnError)

		booking, err := svc.CreateBooking(ctx, validRequest)

		assert.Error(t, err)
		assert.Nil(t, booking)
		assert.Contains(t, err.Error(), "error creating booking")
		mockRepo.AssertExpectations(t)
		mockSpaceX.AssertExpectations(t)
	})

	t.Run("SpaceX conflict", func(t *testing.T) {
		mockRepo := new(mocks.MockBookingRepository)
		mockSpaceX := new(mocks.MockSpaceXClient)
		svc := service.NewBookingService(mockRepo, mockSpaceX)
		ctx := context.Background()

		mockRepo.On("GetDestinationById", ctx, validDestinationID.String()).Return(validDestination, nil)
		mockRepo.On("GetFlights", ctx, mock.Anything).Return([]models.Flight{}, nil)
		mockRepo.On("IsLaunchPadWeekAvailable", ctx, "pad-1", validDestinationID.String(), validLaunchDate).Return(true, nil)
		mockSpaceX.On("CheckLaunchConflict", ctx, "pad-1", validLaunchDate).Return(false, nil)

		booking, err := svc.CreateBooking(ctx, validRequest)

		assert.Error(t, err)
		assert.Nil(t, booking)
		assert.Contains(t, err.Error(), "reserved by SpaceX")
		mockRepo.AssertExpectations(t)
		mockSpaceX.AssertExpectations(t)
	})

	t.Run("SpaceX API not available", func(t *testing.T) {
		validDestinationID := uuid.New()
		validLaunchDate := time.Now().Add(24 * time.Hour)
		validRequest := &models.BookingRequest{
			FirstName:     "John",
			LastName:      "Doe",
			Gender:        "Male",
			Birthday:      time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			LaunchpadID:   "pad-1",
			DestinationID: validDestinationID,
			LaunchDate:    validLaunchDate,
		}
		validDestination := &models.Destination{
			ID:   validDestinationID,
			Name: "Mars",
		}

		mockRepo := new(mocks.MockBookingRepository)
		mockSpaceX := &mocks.MockSpaceXClientUnavailable{}
		svc := service.NewBookingService(mockRepo, mockSpaceX)
		ctx := context.Background()

		mockRepo.On("GetDestinationById", ctx, validDestinationID.String()).Return(validDestination, nil)
		mockRepo.On("GetFlights", ctx, mock.Anything).Return([]models.Flight{}, nil)
		mockRepo.On("IsLaunchPadWeekAvailable", ctx, "pad-1", validDestinationID.String(), validLaunchDate).Return(true, nil)

		booking, err := svc.CreateBooking(ctx, validRequest)

		assert.Error(t, err)
		assert.Nil(t, booking)
		assert.Contains(t, err.Error(), "reserved by SpaceX")
		mockRepo.AssertExpectations(t)
	})

	t.Run("SpaceX API returns error", func(t *testing.T) {
		validDestinationID := uuid.New()
		validLaunchDate := time.Now().Add(24 * time.Hour)
		validRequest := &models.BookingRequest{
			FirstName:     "John",
			LastName:      "Doe",
			Gender:        "Male",
			Birthday:      time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			LaunchpadID:   "pad-1",
			DestinationID: validDestinationID,
			LaunchDate:    validLaunchDate,
		}
		validDestination := &models.Destination{
			ID:   validDestinationID,
			Name: "Mars",
		}

		mockRepo := new(mocks.MockBookingRepository)
		mockSpaceX := &mocks.MockSpaceXClientError{}
		svc := service.NewBookingService(mockRepo, mockSpaceX)
		ctx := context.Background()

		mockRepo.On("GetDestinationById", ctx, validDestinationID.String()).Return(validDestination, nil)
		mockRepo.On("GetFlights", ctx, mock.Anything).Return([]models.Flight{}, nil)
		mockRepo.On("IsLaunchPadWeekAvailable", ctx, "pad-1", validDestinationID.String(), validLaunchDate).Return(true, nil)

		booking, err := svc.CreateBooking(ctx, validRequest)

		assert.Error(t, err)
		assert.Nil(t, booking)
		assert.Contains(t, err.Error(), "error checking SpaceX availability")
		mockRepo.AssertExpectations(t)
	})
}

func TestDeleteBooking(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		mockRepo := new(mocks.MockBookingRepository)
		mockSpaceX := new(mocks.MockSpaceXClient)
		svc := service.NewBookingService(mockRepo, mockSpaceX)

		bookingID := uuid.New().String()
		ctx := context.Background()

		mockBooking := &models.Booking{
			ID:     uuid.MustParse(bookingID),
			Status: models.StatusActive,
		}

		mockRepo.On("GetBookingByID", ctx, bookingID).Return(mockBooking, nil)
		mockRepo.On("DeleteBooking", ctx, bookingID).Return(nil)

		err := svc.DeleteBooking(ctx, bookingID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid UUID", func(t *testing.T) {
		mockRepo := new(mocks.MockBookingRepository)
		mockSpaceX := new(mocks.MockSpaceXClient)
		svc := service.NewBookingService(mockRepo, mockSpaceX)

		err := svc.DeleteBooking(context.Background(), "invalid-uuid")

		assert.Error(t, err)
		assert.Equal(t, models.ErrInvalidUUID, err)
		mockRepo.AssertNotCalled(t, "DeleteBooking")
	})

	t.Run("booking not found", func(t *testing.T) {
		mockRepo := new(mocks.MockBookingRepository)
		mockSpaceX := new(mocks.MockSpaceXClient)
		svc := service.NewBookingService(mockRepo, mockSpaceX)

		bookingID := uuid.New().String()
		ctx := context.Background()

		mockRepo.On("GetBookingByID", ctx, bookingID).Return(nil, models.ErrBookingNotFound)

		err := svc.DeleteBooking(ctx, bookingID)

		assert.Error(t, err)
		assert.Equal(t, models.ErrBookingNotFound, err)
		mockRepo.AssertNotCalled(t, "DeleteBooking")
	})

	t.Run("cannot delete cancelled booking", func(t *testing.T) {
		mockRepo := new(mocks.MockBookingRepository)
		mockSpaceX := new(mocks.MockSpaceXClient)
		svc := service.NewBookingService(mockRepo, mockSpaceX)

		bookingID := uuid.New().String()
		ctx := context.Background()

		mockBooking := &models.Booking{
			ID:     uuid.MustParse(bookingID),
			Status: models.StatusCancelled,
		}

		mockRepo.On("GetBookingByID", ctx, bookingID).Return(mockBooking, nil)

		err := svc.DeleteBooking(ctx, bookingID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete booking with status")
		mockRepo.AssertNotCalled(t, "DeleteBooking")
	})
}

func TestAllBookings(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(mocks.MockBookingRepository)
		mockSpaceX := new(mocks.MockSpaceXClient)
		svc := service.NewBookingService(mockRepo, mockSpaceX)

		ctx := context.Background()
		cursor := "some-cursor"
		limit := 10

		mockBookings := utils.CreateMockBookings(2)
		nextCursor := "next-cursor"

		mockRepo.On("GetBookingsPaginated", ctx, cursor, limit).
			Return(mockBookings, nextCursor, nil)

		getReq := models.GetBookingsRequest{
			Limit: limit,
			Uuid:  cursor,
		}
		response, err := svc.AllBookings(ctx, getReq)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Bookings, 2)
		assert.Equal(t, limit, response.Limit)
		assert.Equal(t, nextCursor, response.Cursor)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(mocks.MockBookingRepository)
		mockSpaceX := new(mocks.MockSpaceXClient)
		svc := service.NewBookingService(mockRepo, mockSpaceX)

		ctx := context.Background()

		// Return empty slice instead of nil for first argument
		mockRepo.On("GetBookingsPaginated", ctx, "", 10).
			Return([]models.Booking{}, "", errors.New("database error"))

		getReq := models.GetBookingsRequest{
			Limit: 10,
			Uuid:  "",
		}
		response, err := svc.AllBookings(ctx, getReq)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "error fetching bookings")
		mockRepo.AssertExpectations(t)
	})

	t.Run("negative limit converted to default", func(t *testing.T) {
		mockRepo := new(mocks.MockBookingRepository)
		mockSpaceX := new(mocks.MockSpaceXClient)
		svc := service.NewBookingService(mockRepo, mockSpaceX)

		ctx := context.Background()

		// Service should convert negative limit to 10 before calling repository
		mockRepo.On("GetBookingsPaginated", ctx, "", 10).
			Return([]models.Booking{}, "", nil)

		getReq := models.GetBookingsRequest{
			Limit: -5,
			Uuid:  "",
		}
		response, err := svc.AllBookings(ctx, getReq)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 10, response.Limit)
		mockRepo.AssertExpectations(t)
	})

	t.Run("zero limit converted to default", func(t *testing.T) {
		mockRepo := new(mocks.MockBookingRepository)
		mockSpaceX := new(mocks.MockSpaceXClient)
		svc := service.NewBookingService(mockRepo, mockSpaceX)

		ctx := context.Background()

		mockRepo.On("GetBookingsPaginated", ctx, "", 10).
			Return([]models.Booking{}, "", nil)

		getReq := models.GetBookingsRequest{
			Limit: 0,
			Uuid:  "",
		}
		response, err := svc.AllBookings(ctx, getReq)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 10, response.Limit)
		mockRepo.AssertExpectations(t)
	})
}
