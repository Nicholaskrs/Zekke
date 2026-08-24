package logtrace

import (
	"sync/atomic"
	"template-go/base/constants"
	"time"

	"github.com/gin-gonic/gin"
)

const traceKey = "_tid"

type LogTrace struct {
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

	// ShouldLogRequest is used to determine whether we need to log Request or not. This can set from anywhere, the goal is
	// to only log request (which can be large) only when necessary (err ,panic or any reason we need the request).
	// The detail itself will only be printed on http logger.
	ShouldLogRequest atomic.Bool
}

func GetLogTrace(c *gin.Context) *LogTrace {
	val, ok := c.Get(traceKey)
	if ok == false {
		// Don't induce panic as much as possible, return an empty struct.
		return &LogTrace{TraceId: "NO_TRACE_ID_FOUND"}
	}

	ret, ok := val.(*LogTrace)
	if ok == false {
		return &LogTrace{TraceId: "INVALID_TRACE_TYPE"}
	}
	return ret
}

func SetLogTrace(c *gin.Context, trace *LogTrace) {
	c.Set(traceKey, trace)
}
