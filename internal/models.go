package models

import (
	"time"

	"github.com/google/uuid"
)

type BookingRequest struct {
	FirstName     string    `json:"first_name" validate:"required,min=2,max=50"`
	LastName      string    `json:"last_name" validate:"required,min=2,max=50"`
	Gender        string    `json:"gender" validate:"required,oneof=Male Female Other"`
	Birthday      time.Time `json:"birthday" validate:"required,valid_age"`
	LaunchpadID   string    `json:"launchpad_id" validate:"required,valid_launchpad"`
	DestinationID uuid.UUID `json:"destination_id" validate:"required,uuid"`
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
