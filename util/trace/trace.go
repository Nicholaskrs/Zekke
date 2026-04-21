package trace

import (
	"template-go/base/constants"
	"time"
)

type Trace struct {
	TraceId string
	Start   time.Time
	Method  constants.MethodType

	// Path is url request path.
	Path string

	// UserId will be filled if user is authenticated.
	UserId int

	// Request is data request. This value might be printed on logger so please ensure
	// to sanitize request if there's secret or any credentials inside also,
	// please sanitize any large request by emptying it or summarize it.
	Request any
}
