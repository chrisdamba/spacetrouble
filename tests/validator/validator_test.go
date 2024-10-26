package validator_test

import (
	"github.com/chrisdamba/spacetrouble/internal/validator"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// Test struct that mimics BookingRequest for validation testing
type TestBooking struct {
	ID            string    `json:"id,omitempty" validate:"omitempty,valid_uuid"`
	FirstName     string    `validate:"required,name_length"`
	LastName      string    `validate:"required,name_length"`
	Gender        string    `validate:"required,gender"`
	Birthday      time.Time `validate:"required,valid_age"`
	LaunchpadID   string    `validate:"required,launchpad_id_length"`
	DestinationID string    `validate:"required,valid_uuid"`
	LaunchDate    time.Time `validate:"required,future_date"`
}

func TestNewCustomValidator(t *testing.T) {
	v := validator.NewCustomValidator()
	assert.NotNil(t, v)
}

func TestValidateFutureDate(t *testing.T) {
	tests := []struct {
		name    string
		booking TestBooking
		wantErr bool
	}{
		{
			name: "Valid future date",
			booking: func() TestBooking {
				b := createValidBaseBooking()
				b.LaunchDate = time.Now().AddDate(0, 1, 0) // One month in future
				return b
			}(),
			wantErr: false,
		},
		{
			name: "Invalid past date",
			booking: func() TestBooking {
				b := createValidBaseBooking()
				b.LaunchDate = time.Now().AddDate(0, -1, 0) // One month in past
				return b
			}(),
			wantErr: true,
		},
	}

	v := validator.NewCustomValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.booking)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAge(t *testing.T) {
	tests := []struct {
		name    string
		booking TestBooking
		wantErr bool
	}{
		{
			name: "Valid age - 20 years old",
			booking: func() TestBooking {
				b := createValidBaseBooking()
				b.Birthday = time.Now().AddDate(-20, 0, 0)
				return b
			}(),
			wantErr: false,
		},
		{
			name: "Invalid age - too young (17)",
			booking: func() TestBooking {
				b := createValidBaseBooking()
				b.Birthday = time.Now().AddDate(-17, 0, 0)
				return b
			}(),
			wantErr: true,
		},
		{
			name: "Invalid age - too old (76)",
			booking: func() TestBooking {
				b := createValidBaseBooking()
				b.Birthday = time.Now().AddDate(-76, 0, 0)
				return b
			}(),
			wantErr: true,
		},
	}

	v := validator.NewCustomValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.booking)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateNameLength(t *testing.T) {
	tests := []struct {
		name    string
		booking TestBooking
		wantErr bool
	}{
		{
			name: "Valid name lengths",
			booking: func() TestBooking {
				b := createValidBaseBooking()
				b.FirstName = "John"
				b.LastName = "Doe"
				return b
			}(),
			wantErr: false,
		},
		{
			name: "Empty first name",
			booking: func() TestBooking {
				b := createValidBaseBooking()
				b.FirstName = ""
				return b
			}(),
			wantErr: true,
		},
		{
			name: "Too long first name",
			booking: func() TestBooking {
				b := createValidBaseBooking()
				b.FirstName = "ThisIsAReallyLongFirstNameThatShouldExceedTheFiftyCharacterLimit"
				return b
			}(),
			wantErr: true,
		},
	}

	v := validator.NewCustomValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.booking)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateGender(t *testing.T) {
	tests := []struct {
		name    string
		booking TestBooking
		wantErr bool
	}{
		{
			name: "Valid gender - male",
			booking: func() TestBooking {
				b := createValidBaseBooking()
				b.Gender = "male"
				return b
			}(),
			wantErr: false,
		},
		{
			name: "Invalid gender",
			booking: func() TestBooking {
				b := createValidBaseBooking()
				b.Gender = "invalid"
				return b
			}(),
			wantErr: true,
		},
		{
			name: "Empty gender",
			booking: func() TestBooking {
				b := createValidBaseBooking()
				b.Gender = ""
				return b
			}(),
			wantErr: true,
		},
	}

	v := validator.NewCustomValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.booking)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateLaunchpadIDLength(t *testing.T) {
	tests := []struct {
		name    string
		booking TestBooking
		wantErr bool
	}{
		{
			name: "Valid launchpad ID length",
			booking: func() TestBooking {
				b := createValidBaseBooking()
				b.LaunchpadID = "123456789012345678901234"
				return b
			}(),
			wantErr: false,
		},
		{
			name: "Invalid launchpad ID - too short",
			booking: func() TestBooking {
				b := createValidBaseBooking()
				b.LaunchpadID = "12345"
				return b
			}(),
			wantErr: true,
		},
	}

	v := validator.NewCustomValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.booking)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateUUID(t *testing.T) {
	validUUID := uuid.New().String()

	tests := []struct {
		name    string
		booking TestBooking
		wantErr bool
	}{
		{
			name: "Valid UUID",
			booking: func() TestBooking {
				b := createValidBaseBooking()
				b.DestinationID = validUUID
				return b
			}(),
			wantErr: false,
		},
		{
			name: "Invalid UUID format",
			booking: func() TestBooking {
				b := createValidBaseBooking()
				b.DestinationID = "not-a-uuid"
				return b
			}(),
			wantErr: true,
		},
	}

	v := validator.NewCustomValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.booking)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCompleteBookingValidation(t *testing.T) {
	validUUID := uuid.New().String()

	tests := []struct {
		name     string
		booking  TestBooking
		wantErr  bool
		errorMsg string
	}{
		{
			name: "Complete valid booking",
			booking: TestBooking{
				FirstName:     "John",
				LastName:      "Doe",
				Gender:        "male",
				Birthday:      time.Now().AddDate(-30, 0, 0),
				LaunchpadID:   "123456789012345678901234",
				DestinationID: validUUID,
				LaunchDate:    time.Now().AddDate(0, 1, 0),
			},
			wantErr: false,
		},
		{
			name: "Multiple validation failures",
			booking: TestBooking{
				FirstName:     "",
				LastName:      "ThisIsAReallyLongLastNameThatShouldExceedTheFiftyCharacterLimit",
				Gender:        "invalid",
				Birthday:      time.Now().AddDate(-80, 0, 0),
				LaunchpadID:   "123",
				DestinationID: "not-a-uuid",
				LaunchDate:    time.Now().AddDate(0, -1, 0),
			},
			wantErr: true,
		},
	}

	v := validator.NewCustomValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.booking)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func createValidBaseBooking() TestBooking {
	return TestBooking{
		FirstName:     "John",
		LastName:      "Doe",
		Gender:        "male",
		Birthday:      time.Now().AddDate(-30, 0, 0),
		LaunchpadID:   "123456789012345678901234",
		DestinationID: uuid.New().String(),
		LaunchDate:    time.Now().AddDate(0, 1, 0),
	}
}
