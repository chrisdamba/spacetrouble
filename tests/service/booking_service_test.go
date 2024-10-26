package service_test

import (
	"context"
	models "github.com/chrisdamba/spacetrouble/internal"
	"github.com/chrisdamba/spacetrouble/internal/service"
	"github.com/chrisdamba/spacetrouble/tests/mocks"
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
