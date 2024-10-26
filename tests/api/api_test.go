package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	models "github.com/chrisdamba/spacetrouble/internal"
	"github.com/chrisdamba/spacetrouble/internal/api"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockBookingService struct {
	mock.Mock
}

func (m *mockBookingService) CreateBooking(ctx context.Context, request *models.BookingRequest) (*models.Booking, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Booking), args.Error(1)
}

func TestBookingHandler_Create(t *testing.T) {
	tests := []struct {
		name          string
		request       *models.BookingRequest
		setupMock     func(*mockBookingService)
		expectedCode  int
		expectedError string
		contentType   string
		acceptHeader  string
	}{
		{
			name: "Success_JSON",
			request: &models.BookingRequest{
				FirstName:     "John",
				LastName:      "Doe",
				Gender:        "male",
				Birthday:      time.Now().AddDate(-30, 0, 0),
				LaunchpadID:   "123456789012345678901234",
				DestinationID: uuid.New(),
				LaunchDate:    time.Now().AddDate(0, 1, 0),
			},
			setupMock: func(m *mockBookingService) {
				m.On("CreateBooking", mock.Anything, mock.AnythingOfType("*models.BookingRequest")).
					Return(&models.Booking{
						ID: uuid.New(),
						User: models.User{
							FirstName: "John",
							LastName:  "Doe",
						},
					}, nil)
			},
			expectedCode: http.StatusCreated,
			contentType:  "application/json",
			acceptHeader: "application/json",
		},
		{
			name: "Success_XML",
			request: &models.BookingRequest{
				FirstName:     "John",
				LastName:      "Doe",
				Gender:        "male",
				Birthday:      time.Now().AddDate(-30, 0, 0),
				LaunchpadID:   "123456789012345678901234",
				DestinationID: uuid.New(),
				LaunchDate:    time.Now().AddDate(0, 1, 0),
			},
			setupMock: func(m *mockBookingService) {
				m.On("CreateBooking", mock.Anything, mock.AnythingOfType("*models.BookingRequest")).
					Return(&models.Booking{
						ID: uuid.New(),
						User: models.User{
							FirstName: "John",
							LastName:  "Doe",
						},
					}, nil)
			},
			expectedCode: http.StatusCreated,
			contentType:  "application/json",
			acceptHeader: "application/xml",
		},
		{
			name:          "Invalid_JSON_Body",
			request:       &models.BookingRequest{},
			setupMock:     func(m *mockBookingService) {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "error json decoding body",
			contentType:   "application/json",
			acceptHeader:  "application/json",
		},
		{
			name: "Validation_Error",
			request: &models.BookingRequest{
				FirstName: "", // Invalid empty name
				LastName:  "Doe",
			},
			setupMock:    func(m *mockBookingService) {},
			expectedCode: http.StatusBadRequest,
			contentType:  "application/json",
			acceptHeader: "application/json",
		},
		{
			name: "Service_Error_InvalidUUID",
			request: &models.BookingRequest{
				FirstName:     "John",
				LastName:      "Doe",
				Gender:        "male",
				Birthday:      time.Now().AddDate(-30, 0, 0),
				LaunchpadID:   "123456789012345678901234",
				DestinationID: uuid.New(),
				LaunchDate:    time.Now().AddDate(0, 1, 0),
			},
			setupMock: func(m *mockBookingService) {
				m.On("CreateBooking", mock.Anything, mock.AnythingOfType("*models.BookingRequest")).
					Return(nil, models.ErrInvalidUUID)
			},
			expectedCode: http.StatusBadRequest,
			contentType:  "application/json",
			acceptHeader: "application/json",
		},
		{
			name: "Service_Error_MissingDestination",
			request: &models.BookingRequest{
				FirstName:     "John",
				LastName:      "Doe",
				Gender:        "male",
				Birthday:      time.Now().AddDate(-30, 0, 0),
				LaunchpadID:   "123456789012345678901234",
				DestinationID: uuid.New(),
				LaunchDate:    time.Now().AddDate(0, 1, 0),
			},
			setupMock: func(m *mockBookingService) {
				m.On("CreateBooking", mock.Anything, mock.AnythingOfType("*models.BookingRequest")).
					Return(nil, models.ErrMissingDestination)
			},
			expectedCode: http.StatusNotFound,
			contentType:  "application/json",
			acceptHeader: "application/json",
		},
		{
			name: "Unsupported_Media_Type",
			request: &models.BookingRequest{
				FirstName: "John",
				LastName:  "Doe",
			},
			setupMock:    func(m *mockBookingService) {},
			expectedCode: http.StatusUnsupportedMediaType,
			contentType:  "application/text",
			acceptHeader: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockBookingService)
			tt.setupMock(mockService)

			handler := api.BookingHandler(mockService)

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/bookings", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", tt.contentType)
			req.Header.Set("Accept", tt.acceptHeader)

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if tt.acceptHeader == "application/json" {
				if tt.expectedError != "" {
					var apiError struct {
						Error string `json:"error"`
					}
					err := json.Unmarshal(rr.Body.Bytes(), &apiError)
					assert.NoError(t, err)
					assert.Contains(t, apiError.Error, tt.expectedError)
				} else if tt.expectedCode == http.StatusCreated {
					var booking models.Booking
					err := json.Unmarshal(rr.Body.Bytes(), &booking)
					assert.NoError(t, err)
					assert.NotEmpty(t, booking.ID)
				}
			} else if tt.acceptHeader == "application/xml" {
				assert.Contains(t, rr.Header().Get("Content-Type"), "application/xml")
			}
			mockService.AssertExpectations(t)
		})
	}
}
