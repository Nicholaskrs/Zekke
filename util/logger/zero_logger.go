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
	logger := &log.Logger
	logger.With().Str(loggerKey, loggerId).Logger()
	return &ZerologLogger{
		log: logger,
	}
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
	log := z.PanicNoTrace().(*ZerologLog)
	log.trace = trace
	return log.Str(traceIdKey, trace.TraceId)
}

func (z *ZerologLogger) Fatal(trace *trace.Trace) Log {
	log := z.FatalNoTrace().(*ZerologLog)
	log.trace = trace
	return log.Str(traceIdKey, trace.TraceId)
}

func (z *ZerologLogger) Error(trace *trace.Trace) Log {
	log := z.ErrorNoTrace().(*ZerologLog)
	log.trace = trace
	return log.Str(traceIdKey, trace.TraceId)
}

func (z *ZerologLogger) Warn(trace *trace.Trace) Log {
	log := z.WarnNoTrace().(*ZerologLog)
	log.trace = trace
	return log.Str(traceIdKey, trace.TraceId)
}

func (z *ZerologLogger) Info(trace *trace.Trace) Log {
	log := z.InfoNoTrace().(*ZerologLog)
	log.trace = trace
	return log.Str(traceIdKey, trace.TraceId)
}

func (z *ZerologLogger) Debug(trace *trace.Trace) Log {
	log := z.DebugNoTrace().(*ZerologLog)
	log.trace = trace
	return log.Str(traceIdKey, trace.TraceId)
}

func (z *ZerologLogger) Trace(trace *trace.Trace) Log {
	log := z.TraceNoTrace().(*ZerologLog)
	log.trace = trace
	return log.Str(traceIdKey, trace.TraceId)
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
		dict:  &ZerologDict{ev},
		log:   ev,
		level: zerolog.FatalLevel,
	}
}

func (z *ZerologLogger) PanicNoTrace() Log {
	ev := z.log.WithLevel(zerolog.PanicLevel)
	return &ZerologLog{
		dict:  &ZerologDict{ev},
		log:   ev,
		level: zerolog.PanicLevel,
	}
}

func (z *ZerologLogger) ErrorNoTrace() Log {
	ev := z.log.Error()
	return &ZerologLog{
		dict:  &ZerologDict{ev},
		log:   ev,
		level: zerolog.ErrorLevel,
	}
}

func (z *ZerologLogger) WarnNoTrace() Log {
	ev := z.log.Warn()
	return &ZerologLog{
		dict:  &ZerologDict{ev},
		log:   ev,
		level: zerolog.WarnLevel,
	}
}

func (z *ZerologLogger) InfoNoTrace() Log {
	ev := z.log.Info()
	return &ZerologLog{
		dict:  &ZerologDict{ev},
		log:   ev,
		level: zerolog.InfoLevel,
	}
}

func (z *ZerologLogger) DebugNoTrace() Log {
	ev := z.log.Debug()
	return &ZerologLog{
		dict:  &ZerologDict{ev},
		log:   ev,
		level: zerolog.DebugLevel,
	}
}

func (z *ZerologLogger) TraceNoTrace() Log {
	ev := z.log.Trace()
	return &ZerologLog{
		dict:  &ZerologDict{ev},
		log:   ev,
		level: zerolog.TraceLevel,
	}
}

func (z *ZerologLogger) NewDict() LogDict {
	return &ZerologDict{
		dict: zerolog.Dict(),
	}
}

type ZerologLog struct {
	dict  *ZerologDict
	log   *zerolog.Event
	level zerolog.Level
	trace *trace.Trace
	err   error
}

func (z *ZerologLog) Msg(msg string) {
	z.log.Msg(msg)
}

func (z *ZerologLog) Msgf(msg string, args ...interface{}) {
	z.log.Msgf(msg, args...)
}

func (z *ZerologLog) MarshalJson(key string, data interface{}) Log {
	z.dict.MarshalJson(key, data)
	return z
}

func (z *ZerologLog) RawJson(key string, jsonBytes []byte) Log {
	z.dict.RawJson(key, jsonBytes)
	return z
}

func (z *ZerologLog) Error(err error) Log {
	z.dict.ErrorCustomKey(errKey, err)
	z.err = err
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
	z.dict.Str(key, str)
	return z
}

func (z *ZerologLog) Strs(key string, args ...string) Log {
	z.dict.Strs(key, args...)
	return z
}

func (z *ZerologLog) Bool(key string, val bool) Log {
	z.dict.Bool(key, val)
	return z
}

func (z *ZerologLog) Int(key string, val int) Log {
	z.dict.Int(key, val)
	return z
}

func (z *ZerologLog) Ints(key string, args ...int) Log {
	z.dict.Ints(key, args...)
	return z
}

func (z *ZerologLog) Int64(key string, val int64) Log {
	z.dict.Int64(key, val)
	return z
}

func (z *ZerologLog) Float64(key string, val float64) Log {
	z.dict.Float64(key, val)
	return z
}

func (z *ZerologLog) Bytes(key string, val []byte) Log {
	z.dict.Bytes(key, val)
	return z
}

func (z *ZerologLog) Dict(key string, dict LogDict) Log {
	z.dict.Dict(key, dict)
	return z
}

func (z *ZerologLog) Time(key string, t time.Time) Log {
	z.dict.Time(key, t)
	return z
}

func (z *ZerologLog) Dur(key string, d time.Duration) Log {
	z.dict.Dur(key, d)
	return z
}

type ZerologDict struct {
	dict *zerolog.Event
}

func (z *ZerologDict) MarshalJson(key string, data interface{}) LogDict {
	bytes, parseErr := json.Marshal(data)
	if parseErr != nil {
		z.ErrorCustomKey(key+"parse", parseErr)
		return z
	}
	// Otherwise add the bytes.
	z.dict.RawJSON(key, bytes)
	return z
}

func (z *ZerologDict) RawJson(key string, data []byte) LogDict {
	z.dict.RawJSON(key, data)
	return z
}

func (z *ZerologDict) Error(err error) LogDict {
	return z.ErrorCustomKey(errKey, err)
}

func (z *ZerologDict) PanicError(pErr error) LogDict {
	return z.Error(pErr)
}

func (z *ZerologDict) ErrorCustomKey(key string, err error) LogDict {
	errList := zerolog.Arr()
	level := 0
	for level <= 5 {
		if err == nil {
			break
		}

		// If the error is of LogValues type, we add all of their values to this dictionary.
		if p, ok := err.(LogValues); ok {
			dict := zerolog.Dict()
			p.LogValues(&ZerologDict{
				dict: dict,
			})
			errList.Dict(dict)
			level++

			err = errors.Unwrap(err)
			continue
		}

		// Otherwise we assume nothing and just log the error string.
		errList.Str(err.Error())

		// Unwrap to log the next value.
		err = errors.Unwrap(err)
		level++
	}
	z.dict.Array(key, errList)
	return z
}

func (z *ZerologDict) Str(key string, val string) LogDict {
	z.dict.Str(key, val)
	return z
}

func (z *ZerologDict) Strs(key string, args ...string) LogDict {
	z.dict.Strs(key, args)
	return z
}

func (z *ZerologDict) Bool(key string, val bool) LogDict {
	z.dict.Bool(key, val)
	return z
}

func (z *ZerologDict) Int(key string, val int) LogDict {
	z.dict.Int(key, val)
	return z
}

func (z *ZerologDict) Ints(key string, args ...int) LogDict {
	z.dict.Ints(key, args)
	return z
}

func (z *ZerologDict) Int64(key string, val int64) LogDict {
	z.dict.Int64(key, val)
	return z
}

func (z *ZerologDict) Float64(key string, val float64) LogDict {
	z.dict.Float64(key, val)
	return z
}

func (z *ZerologDict) Dict(key string, dict LogDict) LogDict {
	d := dict.(*ZerologDict)
	z.dict.Dict(key, d.dict)
	return z
}

func (z *ZerologDict) Bytes(key string, val []byte) LogDict {
	z.dict.Bytes(key, val)
	return z
}

func (z *ZerologDict) Time(key string, t time.Time) LogDict {
	z.dict.Time(key, t)
	return z
}

func (z *ZerologDict) Dur(key string, d time.Duration) LogDict {
	z.dict.Dur(key, d)
	return z
}

func (z *ZerologDict) NewDict() LogDict {
	return &ZerologDict{
		dict: zerolog.Dict(),
	}
}
