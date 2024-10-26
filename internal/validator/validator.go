package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"time"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	v := validator.New()
	v.RegisterValidation("gender", validateGender)
	v.RegisterValidation("valid_uuid", validateUUID)
	v.RegisterValidation("future_date", validateFutureDate)
	v.RegisterValidation("valid_age", validateAge)
	v.RegisterValidation("valid_destination", validateDestination)
	v.RegisterValidation("name_length", validateNameLength)
	v.RegisterValidation("valid_launchpad", validateLaunchpad)
	v.RegisterValidation("launchpad_id_length", validateLaunchpadIDLength)

	return &CustomValidator{validator: v}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func validateFutureDate(fl validator.FieldLevel) bool {
	date, ok := fl.Field().Interface().(time.Time)
	if !ok {
		return false
	}
	return date.After(time.Now())
}

func validateAge(fl validator.FieldLevel) bool {
	birthday, ok := fl.Field().Interface().(time.Time)
	if !ok {
		return false
	}
	age := time.Now().Year() - birthday.Year()
	return age >= 18 && age <= 75
}

func validateDestination(fl validator.FieldLevel) bool {
	destID := fl.Field().Interface().(int)
	validDestinations := map[int]bool{
		1: true, // Mars
		2: true, // Moon
		3: true, // Pluto
		4: true, // Asteroid Belt
		5: true, // Europa
		6: true, // Titan
		7: true, // Ganymede
	}
	return validDestinations[destID]
}

func validateLaunchpad(fl validator.FieldLevel) bool {
	launchpadID := fl.Field().Interface().(int)
	return launchpadID > 0 && launchpadID <= 5
}

func validateNameLength(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	return len(name) > 0 && len(name) <= 50
}

func validateGender(fl validator.FieldLevel) bool {
	gender := fl.Field().String()
	supportedGenders := map[string]bool{
		"female": true,
		"male":   true,
		"other":  true,
	}
	return supportedGenders[gender]
}

func validateLaunchpadIDLength(fl validator.FieldLevel) bool {
	launchpadID := fl.Field().String()
	return len(launchpadID) == 24
}

func validateUUID(fl validator.FieldLevel) bool {
	id := fl.Field().String()
	_, err := uuid.Parse(id)
	return err == nil
}
