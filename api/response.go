package api

import (
	"fmt"
	"net/http"
	"template-go/base/constants"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Response struct {
	TraceID          string            `json:"trace_id"`
	Message          string            `json:"message"`
	Data             interface{}       `json:"data"`
	Meta             *PaginationMeta   `json:"meta,omitempty"`
	ValidationErrors []ValidationError `json:"errors"`
}

// ReturnValidationError used to return json response with validation errors
func ReturnValidationError(c *gin.Context, module string, messageType constants.MethodType, data interface{}, err validator.ValidationErrors, traceID string) {
	var messagePrefix string
	if messageType != "" {
		messagePrefix = fmt.Sprintf("%s %s", messageType, module)
	} else {
		messagePrefix = module
	}
	c.JSON(http.StatusUnprocessableEntity, Response{
		TraceID:          traceID,
		Message:          fmt.Sprintf("%s failed", messagePrefix),
		ValidationErrors: SetValidationMessage(err, data),
	})
}

// ReturnResponse used to return json response. It will check response status.
// If response status == OK, the it will return data wil http status 200.
// Otherwise, it wull return based on parameter status code and return failed message.
func ReturnResponse(c *gin.Context, module string, messageType constants.MethodType, data interface{}, message string, statusCode int, traceID string) {
	var messagePrefix string
	if messageType != "" {
		messagePrefix = fmt.Sprintf("%s %s", messageType, module)
	} else {
		messagePrefix = module
	}
	if statusCode == http.StatusOK {
		data := Response{
			TraceID: traceID,
			Message: fmt.Sprintf("%s success", messagePrefix),
			Data:    data,
		}
		c.JSON(http.StatusOK, data)
		return
	}

	c.JSON(statusCode, Response{
		TraceID: traceID,
		Message: fmt.Sprintf("%s failed: %s", messagePrefix, message),
	})

}

// ReturnInternalServerError used to return json response with http error code 500.
func ReturnInternalServerError(c *gin.Context, module string, messageType constants.MethodType, err error, traceID string) {
	var messagePrefix string
	if messageType != "" {
		messagePrefix = fmt.Sprintf("%s %s", messageType, module)
	} else {
		messagePrefix = module
	}
	c.JSON(http.StatusInternalServerError, Response{
		TraceID: traceID,
		Message: fmt.Sprintf("%s failed: %s", messagePrefix, err.Error()),
	})

}
