package trace

import (
	"template-go/util/logtrace"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GinHandler inserts LogTrace object to the Gin context.
func GinHandler(c *gin.Context) {
	startTime := time.Now()

	ctx, span := Tracer("zekke").Start(c.Request.Context(), c.FullPath())
	defer span.End()

	c.Request = c.Request.WithContext(ctx)
	spanCtx := span.SpanContext()
	traceId := ""
	if spanCtx.IsValid() {
		traceId = spanCtx.TraceID().String()
	} else {
		traceId = uuid.New().String() // fallback so logs still correlate even with OTel off
	}

	trace := logtrace.LogTrace{
		TraceId: traceId,
		Start:   startTime,
		Path:    c.Request.RequestURI,
	}

	// Set the trace object.
	logtrace.SetLogTrace(c, &trace)
	c.Next()
}
