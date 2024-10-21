package router

import (
	"github.com/go-playground/validator/v10"
)

// Validator is the struct that contains the validator.
type Validator struct {
	Vdt *validator.Validate
}

// NewValidator creates a new Validator.
func NewValidator() (*Validator, error) {
	vdt := validator.New()
	return &Validator{
		Vdt: vdt,
	}, nil
}

// Validate validates a struct using the validator.
func (v *Validator) Validate(i interface{}) error {
	return v.Vdt.Struct(i)
}
