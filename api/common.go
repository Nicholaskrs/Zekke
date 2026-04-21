package api

import (
	"fmt"
	"net/http"
	"template-go/base/constants"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ReturnValidationError used to return json response with validation errors
func ReturnValidationError(c *gin.Context, module string, messageType constants.MethodType, data interface{}, err validator.ValidationErrors, traceID string) {
	var messagePrefix string
	if messageType != "" {
		messagePrefix = fmt.Sprintf("%s %s", messageType, module)
	} else {
		messagePrefix = module
	}
	c.JSON(http.StatusUnprocessableEntity, ApiResponse{
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
		data := ApiResponse{
			TraceID: traceID,
			Message: fmt.Sprintf("%s success", messagePrefix),
			Data:    data,
		}
		c.JSON(http.StatusOK, data)
		return
	}

	c.JSON(statusCode, ApiResponse{
		TraceID: traceID,
		Message: fmt.Sprintf("%s failed: %s", messagePrefix, message),
	})

}

// ReturnInternalServerError used to return json response with http error code 500.
func ReturnInternalServerError(c *gin.Context, module string, messageType constants.MethodType, data interface{}, err error, traceID string) {
	var messagePrefix string
	if messageType != "" {
		messagePrefix = fmt.Sprintf("%s %s", messageType, module)
	} else {
		messagePrefix = module
	}
	c.JSON(http.StatusInternalServerError, ApiResponse{
		TraceID: traceID,
		Message: fmt.Sprintf("%s failed: %s", messagePrefix, err.Error()),
	})

}

// ReturnResponsePagination used to return json response with meta data for pagination. It will check response status.
// If response status == OK, the it will return data wil http status 200.
// Otherwise, it wull return based on parameter status code and return failed message.
func ReturnResponsePagination(c *gin.Context, module string, messageType constants.MethodType, data interface{}, message string, statusCode int, traceID string, meta *MetaResponse) {
	var messagePrefix string
	if messageType != "" {
		messagePrefix = fmt.Sprintf("%s %s", messageType, module)
	} else {
		messagePrefix = module
	}
	if statusCode == http.StatusOK {
		data := ApiResponse{
			TraceID: traceID,
			Message: fmt.Sprintf("%s success", messagePrefix),
			Data:    data,
			Meta:    meta,
		}
		c.JSON(http.StatusOK, data)
		return
	}

	c.JSON(statusCode, ApiResponse{
		TraceID: traceID,
		Message: fmt.Sprintf("%s failed: %s", messagePrefix, message),
		Meta:    &MetaResponse{},
	})

}
