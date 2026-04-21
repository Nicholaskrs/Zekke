package api

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
)

type ApiResponse struct {
	TraceID          string            `json:"trace_id"`
	Message          string            `json:"message"`
	Data             interface{}       `json:"data"`
	Meta             *MetaResponse     `json:"meta,omitempty"`
	ValidationErrors []ValidationError `json:"errors"`
}

// MetaResponse used for pagination. Will be set into nil if there is no pagination in the BaseResponse.Data.
type MetaResponse struct {
	Page      int   `json:"page"`
	Limit     int   `json:"limit"`
	TotalRows int64 `json:"totalRows"`
}

type ValidationError struct {
	Field string `json:"field"`
	Msg   string `json:"message"`
}

type PaginateRequest struct {
	Page  int `form:"page" binding:"required,min=1"`
	Limit int `form:"limit" binding:"required,min=1,max=100"`
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
