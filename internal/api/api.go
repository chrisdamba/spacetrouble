package api

import (
	models "github.com/chrisdamba/spacetrouble/internal"
	"github.com/chrisdamba/spacetrouble/internal/ports"
	"github.com/chrisdamba/spacetrouble/internal/utils"
	"github.com/chrisdamba/spacetrouble/internal/validator"
	"github.com/google/uuid"
	"net/http"
	"strconv"
)

func BookingHandler(service ports.BookingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			create(service, w, r)
		case http.MethodDelete:
			deleteBooking(service, w, r)
		case http.MethodGet:
			list(service, w, r)
		}
	}
}

func create(service ports.BookingService, w http.ResponseWriter, r *http.Request) {
	var bookingRequest models.BookingRequest
	if err := utils.JsonDecodeBody(r, &bookingRequest); err != nil {
		ae := utils.NewBadRequest("error json decoding body")
		utils.RenderResponse(r, w, ae.StatusCode, ae)
		return
	}

	v := validator.NewCustomValidator()
	if err := v.Validate(bookingRequest); err != nil {
		ae := utils.NewBadRequest(err.Error())
		utils.RenderResponse(r, w, ae.StatusCode, ae)
		return
	}

	ans, err := service.CreateBooking(r.Context(), &bookingRequest)
	if err != nil {
		ae := getApiError(err)
		utils.RenderResponse(r, w, ae.StatusCode, ae)
		return

	}
	utils.RenderResponse(r, w, http.StatusCreated, ans)
}

func list(service ports.BookingService, w http.ResponseWriter, r *http.Request) {
	cursor := r.URL.Query().Get("cursor")
	limitStr := r.URL.Query().Get("limit")

	limit := 10 // default limit
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 {
			ae := utils.NewBadRequest("invalid limit parameter")
			utils.RenderResponse(r, w, ae.StatusCode, ae)
			return
		}
		limit = parsedLimit
	}

	getReq := models.GetBookingsRequest{
		Limit: limit,
	}
	if cursor != "" {
		var err error
		var getReqUuid uuid.UUID
		_, getReqUuid, err = utils.DecodeCursor(cursor)
		if err != nil {
			ae := utils.NewBadRequest(err.Error())
			utils.RenderResponse(r, w, ae.StatusCode, ae)
			return
		}
		getReq.Uuid = getReqUuid.String()
	}
	bookings, err := service.AllBookings(r.Context(), getReq)
	if err != nil {
		ae := getApiError(err)
		utils.RenderResponse(r, w, ae.StatusCode, ae)
		return
	}

	utils.RenderResponse(r, w, http.StatusOK, bookings)
}

func deleteBooking(service ports.BookingService, w http.ResponseWriter, r *http.Request) {
	bookingID := r.URL.Query().Get("id")
	if bookingID == "" {
		ae := utils.NewBadRequest("booking ID is required")
		utils.RenderResponse(r, w, ae.StatusCode, ae)
		return
	}

	if err := service.DeleteBooking(r.Context(), bookingID); err != nil {
		ae := getApiError(err)
		utils.RenderResponse(r, w, ae.StatusCode, ae)
		return
	}

	utils.RenderResponse(r, w, http.StatusNoContent, nil)
}

func getApiError(err error) utils.ApiError {
	ae := utils.ApiError{Msg: err.Error()}
	switch err {
	case models.ErrInvalidUUID:
		ae.StatusCode = http.StatusBadRequest
	case models.ErrMissingDestination:
		ae.StatusCode = http.StatusNotFound
	case models.ErrBookingNotFound:
		ae.StatusCode = http.StatusNotFound
	case models.ErrLaunchPadUnavailable:
		ae.StatusCode = http.StatusConflict
	default:
		ae.StatusCode = http.StatusInternalServerError
	}
	return ae
}
