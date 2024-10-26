package models

import (
	"time"

	"github.com/google/uuid"
)

type BookingRequest struct {
	ID            string    `json:"id,omitempty" validate:"omitempty,valid_uuid"`
	FirstName     string    `json:"first_name" validate:"required,name_length"`
	LastName      string    `json:"last_name" validate:"required,name_length"`
	Gender        string    `json:"gender" validate:"required,gender"`
	Birthday      time.Time `json:"birthday" validate:"required,valid_age"`
	LaunchpadID   string    `json:"launchpad_id" validate:"required,launchpad_id_length"`
	DestinationID uuid.UUID `json:"destination_id" validate:"required,valid_uuid"`
	LaunchDate    time.Time `json:"launch_date" validate:"required,future_date"`
}

type AllBookingsResponse struct {
	Bookings []BookingResponse `json:"bookings"`
	Limit    int               `json:"limit"`
	Cursor   string            `json:"cursor"`
}

type GetBookingsRequest struct {
	Limit int
	Uuid  string
	Ts    time.Time
}

type BookingStatus string

const (
	StatusActive    BookingStatus = "ACTIVE"
	StatusConfirmed BookingStatus = "CONFIRMED"
	StatusCancelled BookingStatus = "CANCELLED"
)

type Destination struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type Flight struct {
	ID          uuid.UUID   `json:"id"`
	LaunchpadID string      `json:"launchpad_id"`
	Destination Destination `json:"destination"`
	LaunchDate  time.Time   `json:"launch_date"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Gender    string    `json:"gender"`
	Birthday  time.Time `json:"birthday"`
}

type Booking struct {
	ID        uuid.UUID     `json:"id"`
	User      User          `json:"user"`
	Flight    Flight        `json:"flight"`
	Status    BookingStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
}

type BookingResponse struct {
	Booking
}
