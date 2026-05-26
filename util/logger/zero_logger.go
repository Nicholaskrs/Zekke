package logger

import (
	"encoding/json"
	"errors"
	"fmt"
	"template-go/util/trace"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/gin-gonic/gin"
)

const (
	errKey     = "_err"
	traceIdKey = "_traceid"
	loggerKey  = "_logger"
)

var _ Logger = (*ZerologLogger)(nil)
var _ Log = (*ZerologLog)(nil)

func NewZerologLogger(loggerId string) *ZerologLogger {
	logger := log.Logger.With().Str(loggerKey, loggerId).Logger()
	return &ZerologLogger{log: &logger}
}

type ZerologLogger struct {
	log *zerolog.Logger
}

func (*ZerologLogger) RouterLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		log.Printf("http: %s %s (%3d) [%v]",
			method,
			path,
			statusCode,
			latency,
		)
	}
}

func (z *ZerologLogger) Panic(trace *trace.Trace) Log {
	return z.PanicNoTrace().Str(traceIdKey, trace.TraceId)
}

func (z *ZerologLogger) Fatal(trace *trace.Trace) Log {
	return z.FatalNoTrace().Str(traceIdKey, trace.TraceId)
}

func (z *ZerologLogger) Error(trace *trace.Trace) Log {
	return z.ErrorNoTrace().Str(traceIdKey, trace.TraceId)
}

func (z *ZerologLogger) Warn(trace *trace.Trace) Log {
	return z.WarnNoTrace().Str(traceIdKey, trace.TraceId)
}

func (z *ZerologLogger) Info(trace *trace.Trace) Log {
	return z.InfoNoTrace().Str(traceIdKey, trace.TraceId)
}

func (z *ZerologLogger) Debug(trace *trace.Trace) Log {
	return z.DebugNoTrace().Str(traceIdKey, trace.TraceId)
}

func (z *ZerologLogger) Trace(trace *trace.Trace) Log {
	return z.TraceNoTrace().Str(traceIdKey, trace.TraceId)
}

func (z *ZerologLogger) PanicErr(trace *trace.Trace, err error) Log {
	return z.Panic(trace).Error(err)
}

func (z *ZerologLogger) FatalErr(trace *trace.Trace, err error) Log {
	return z.Fatal(trace).Error(err)
}

func (z *ZerologLogger) ErrorErr(trace *trace.Trace, err error) Log {
	return z.Error(trace).Error(err)
}

func (z *ZerologLogger) WarnErr(trace *trace.Trace, err error) Log {
	return z.Warn(trace).Error(err)
}

func (z *ZerologLogger) InfoErr(trace *trace.Trace, err error) Log {
	return z.Info(trace).Error(err)
}

func (z *ZerologLogger) DebugErr(trace *trace.Trace, err error) Log {
	return z.Debug(trace).Error(err)
}

func (z *ZerologLogger) TraceErr(trace *trace.Trace, err error) Log {
	return z.Trace(trace).Error(err)
}

func (z *ZerologLogger) FatalNoTrace() Log {
	ev := z.log.WithLevel(zerolog.FatalLevel)
	return &ZerologLog{
		log:   ev,
		level: zerolog.FatalLevel,
	}
}

func (z *ZerologLogger) PanicNoTrace() Log {
	ev := z.log.WithLevel(zerolog.PanicLevel)
	return &ZerologLog{
		log:   ev,
		level: zerolog.PanicLevel,
	}
}

func (z *ZerologLogger) ErrorNoTrace() Log {
	ev := z.log.Error()
	return &ZerologLog{
		log:   ev,
		level: zerolog.ErrorLevel,
	}
}

func (z *ZerologLogger) WarnNoTrace() Log {
	ev := z.log.Warn()
	return &ZerologLog{
		log:   ev,
		level: zerolog.WarnLevel,
	}
}

func (z *ZerologLogger) InfoNoTrace() Log {
	ev := z.log.Info()
	return &ZerologLog{
		log:   ev,
		level: zerolog.InfoLevel,
	}
}

func (z *ZerologLogger) DebugNoTrace() Log {
	ev := z.log.Debug()
	return &ZerologLog{
		log:   ev,
		level: zerolog.DebugLevel,
	}
}

func (z *ZerologLogger) TraceNoTrace() Log {
	ev := z.log.Trace()
	return &ZerologLog{
		log:   ev,
		level: zerolog.TraceLevel,
	}
}

type ZerologLog struct {
	log   *zerolog.Event
	level zerolog.Level
}

func (z *ZerologLog) Msg(msg string) {
	z.log.Msg(msg)
}

func (z *ZerologLog) Msgf(msg string, args ...interface{}) {
	z.log.Msgf(msg, args...)
}

func (z *ZerologLog) MarshalJson(key string, data interface{}) Log {
	bytes, parseErr := json.Marshal(data)
	if parseErr != nil {
		z.Error(parseErr).Str(key, fmt.Sprintf("Failed to marshal data: %v", parseErr))
		return z
	}
	z.log.RawJSON(key, bytes)
	return z
}

func (z *ZerologLog) RawJson(key string, jsonBytes []byte) Log {
	z.log.RawJSON(key, jsonBytes)
	return z
}

func (z *ZerologLog) Error(err error) Log {
	z.log.Str(errKey, err.Error())
	return z
}

func (z *ZerologLog) PanicError(pErr interface{}) Log {
	err, ok := pErr.(error)
	if ok == false {
		// Else treat as a value. Create an error.
		err = errors.New(fmt.Sprintf("%v", pErr))
	}
	return z.Error(err)
}

func (z *ZerologLog) Str(key string, str string) Log {
	z.log.Str(key, str)
	return z
}

func (z *ZerologLog) Strs(key string, args ...string) Log {
	z.log.Strs(key, args)
	return z
}

func (z *ZerologLog) Bool(key string, val bool) Log {
	z.log.Bool(key, val)
	return z
}

func (z *ZerologLog) Int(key string, val int) Log {
	z.log.Int(key, val)
	return z
}

func (z *ZerologLog) Ints(key string, args ...int) Log {
	z.log.Ints(key, args)
	return z
}

func (z *ZerologLog) Int64(key string, val int64) Log {
	z.log.Int64(key, val)
	return z
}

func (z *ZerologLog) Float64(key string, val float64) Log {
	z.log.Float64(key, val)
	return z
}

func (z *ZerologLog) Bytes(key string, val []byte) Log {
	z.log.Bytes(key, val)
	return z
}

func (z *ZerologLog) Time(key string, t time.Time) Log {
	z.log.Time(key, t)
	return z
}

func (z *ZerologLog) Dur(key string, d time.Duration) Log {
	z.log.Dur(key, d)
	return z
}
