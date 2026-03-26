package validator

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Validate validates a struct against its validation tags.
func Validate(s interface{}) error {
	return validate.Struct(s)
}
