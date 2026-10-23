package api

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Field string `json:"field"`
	Msg   string `json:"message"`
}

func ValidationMessage(tag string, param string) string {
	switch tag {
	case "required":
		return "The field is required"
	case "email":
		return "Invalid email"
	case "min":
		return "The value below the limit"
	case "max":
		return "The value exceeded the limit"
	case "lte":
		return fmt.Sprintf("The value must be less than equal %s", param)
	case "gte":
		return fmt.Sprintf("The value must be more than equal %s", param)
	case "oneof":
		values := strings.ReplaceAll(param, " ", "', '")
		return fmt.Sprintf("The value must be one of ['%s']", values)
	}
	return ""
}

func SetValidationMessage(ve validator.ValidationErrors, data interface{}) []ValidationError {
	out := make([]ValidationError, 0)
	for _, fe := range ve {
		fieldName := fe.Field()
		typeOf := reflect.TypeOf(data).Elem()
		field, _ := typeOf.FieldByName(fieldName)
		fieldJSONName, _ := field.Tag.Lookup("json")
		if fieldJSONName == "" {
			fieldJSONName, _ = field.Tag.Lookup("form")
		}
		if fieldJSONName == "" {
			fieldJSONName, _ = field.Tag.Lookup("uri")
		}
		out = append(out, ValidationError{
			Field: fieldJSONName,
			Msg:   ValidationMessage(fe.Tag(), fe.Param()),
		})
	}
	return out
}
