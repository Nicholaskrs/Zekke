package logger

import (
	"template-go/util/trace"
	"time"

	"github.com/gin-gonic/gin"
)

type Logger interface {
	RouterLogger() gin.HandlerFunc

	PanicNoTrace() Log
	FatalNoTrace() Log
	ErrorNoTrace() Log
	WarnNoTrace() Log
	InfoNoTrace() Log
	DebugNoTrace() Log
	TraceNoTrace() Log

	Panic(trace *trace.Trace) Log
	Fatal(trace *trace.Trace) Log
	Error(trace *trace.Trace) Log
	Warn(trace *trace.Trace) Log
	Info(trace *trace.Trace) Log
	Debug(trace *trace.Trace) Log
	Trace(trace *trace.Trace) Log

	PanicErr(trace *trace.Trace, err error) Log
	FatalErr(trace *trace.Trace, err error) Log
	ErrorErr(trace *trace.Trace, err error) Log
	WarnErr(trace *trace.Trace, err error) Log
	InfoErr(trace *trace.Trace, err error) Log
	DebugErr(trace *trace.Trace, err error) Log
	TraceErr(trace *trace.Trace, err error) Log
}

type Log interface {
	MarshalJson(key string, data interface{}) Log
	RawJson(key string, jsonBytes []byte) Log
	Error(err error) Log
	PanicError(pErr interface{}) Log
	Str(key string, str string) Log
	Strs(key string, args ...string) Log
	Bool(key string, val bool) Log
	Int(key string, val int) Log
	Ints(key string, args ...int) Log
	Int64(key string, val int64) Log
	Float64(key string, val float64) Log
	Bytes(key string, val []byte) Log
	Time(key string, t time.Time) Log
	Dur(key string, d time.Duration) Log

	// Msg prints the log with the given message.
	Msg(msg string)
}
