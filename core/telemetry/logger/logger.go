package logger

import (
	"template-go/util/logtrace"
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

	Panic(trace *logtrace.LogTrace) Log
	Fatal(trace *logtrace.LogTrace) Log
	Error(trace *logtrace.LogTrace) Log
	Warn(trace *logtrace.LogTrace) Log
	Info(trace *logtrace.LogTrace) Log
	Debug(trace *logtrace.LogTrace) Log
	Trace(trace *logtrace.LogTrace) Log

	PanicErr(trace *logtrace.LogTrace, err error) Log
	FatalErr(trace *logtrace.LogTrace, err error) Log
	ErrorErr(trace *logtrace.LogTrace, err error) Log
	WarnErr(trace *logtrace.LogTrace, err error) Log
	InfoErr(trace *logtrace.LogTrace, err error) Log
	DebugErr(trace *logtrace.LogTrace, err error) Log
	TraceErr(trace *logtrace.LogTrace, err error) Log
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
	Detail(trace *logtrace.LogTrace) Log

	// Msg prints the log with the given message.
	Msg(msg string)
}
