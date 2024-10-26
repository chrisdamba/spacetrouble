package api

import (
	models "github.com/chrisdamba/spacetrouble/internal"
	"github.com/chrisdamba/spacetrouble/internal/ports"
	"github.com/chrisdamba/spacetrouble/internal/utils"
	"github.com/chrisdamba/spacetrouble/internal/validator"
	"net/http"
)

func BookingHandler(service ports.BookingService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			create(service, w, r)
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

func getApiError(err error) utils.ApiError {
	ae := utils.ApiError{Msg: err.Error()}
	switch err {
	case models.ErrInvalidUUID:
		ae.StatusCode = http.StatusBadRequest
	case models.ErrMissingDestination:
		ae.StatusCode = http.StatusNotFound
	case models.ErrLaunchPadUnavailable:
		ae.StatusCode = http.StatusConflict
	default:
		ae.StatusCode = http.StatusInternalServerError
	}
	return ae
}
