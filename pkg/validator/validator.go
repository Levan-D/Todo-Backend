package validator

import (
	"github.com/go-playground/validator"
)

type ValidateError struct {
	Errors []*ValidateItem `json:"errors"`
}

type ValidateItem struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

func Validate(st interface{}) []*ValidateItem {
	var errors []*ValidateItem
	validate := validator.New()
	err := validate.Struct(st)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, &ValidateItem{
				Field: err.Field(),
				Tag:   err.Tag(),
				Value: err.Param(),
			})
		}
	}
	return *&errors
}
