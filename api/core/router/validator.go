package router

import (
	"github.com/go-playground/validator/v10"
)

// Validator is the struct that contains the validator
type Validator struct {
	validator *validator.Validate
}

// NewValidator creates a new Validator
func NewValidator() (*Validator, error) {
	vdt := validator.New()

	return &Validator{
		validator: vdt,
	}, nil
}

func (v *Validator) RegisterValidator(tag string, fn validator.Func) error {
	return v.validator.RegisterValidation(tag, fn)
}

// Validate validates a struct using the validator
func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}
