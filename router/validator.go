// Copyright © 2024 Ingka Holding B.V. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
