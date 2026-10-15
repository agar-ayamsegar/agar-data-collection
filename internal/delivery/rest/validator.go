package rest

import "github.com/go-playground/validator/v10"

type echoValidator struct {
	v *validator.Validate
}

func (e *echoValidator) Validate(i interface{}) error {
	return e.v.Struct(i)
}
