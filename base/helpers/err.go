package helpers

import (
	"fmt"
	"runtime/debug"
	"strings"
	"template-go/util/logger"
)

var _ error = (*AppErr)(nil)
var _ logger.LogValues = (*AppErr)(nil)

const (
	errMsgKey         = "_msg"
	errStackKey       = "_stack"
	errDebugValuesKey = "_debug"
)

// AppErr is
//
// See: https://blog.golang.org/errors-are-values
// See: https://blog.golang.org/go1.13-errors
type AppErr struct {
	// ErrCode can be used to convey and differentiate between different types of errors.
	ErrCode int

	msg         string
	debugValues map[string]string
	stackTrace  []string
	wrapped     error
}

func (e *AppErr) Error() string {
	msg := e.msg
	if e.wrapped != nil && e.wrapped.Error() != msg {
		msg += " caused by: " + e.wrapped.Error()
	}

	// Possible perf issue if the keys and values are long.
	// However, Error() is rarely called, and we must not use AppError.Str for large values anyway.
	for k, v := range e.debugValues {
		msg += fmt.Sprintf("; key:%v value:%v", k, v)
	}
	return e.msg
}

// Str mimics zerolog.Logger.Str(). Attach the value into this error, allowing this error carry the value upstream
// where it can be logged. Our default logger will check and print these values using zerolog.
//
// Values will be converted to string first before stored. Conversion to string incurs cost, so be sure to only pass
// primitive variables (int, float, string, bool, etc). Don't pass structs or interfaces here.
//
// USE SPARINGLY. DO NOT PASS LARGE VARIABLES USING THIS METHOD. Rule of the thumb, only pass what you would include in
// the error message through this variable. Especially don't pass big structs, objects with connections, etc, through
// this.
func (e *AppErr) Str(key string, val interface{}) *AppErr {
	e.debugValues[key] = fmt.Sprintf("%v", val)
	return e
}

func (e *AppErr) DebugValues() map[string]string {
	return e.debugValues
}

// Msg overwrites the error message of this error. Useful when wrapping third party errors.
// The current message will be inserted to debugValues.
func (e *AppErr) Msg(msg string) *AppErr {
	e.debugValues["__overwritten"] = e.msg
	e.msg = msg
	return e
}

// Code sets the error code for this error. Useful to differentiate between different error types.
func (e *AppErr) Code(errCode int) *AppErr {
	e.ErrCode = errCode
	return e
}

// StackTrace might return nil if AppErr is created without it.
func (e *AppErr) StackTrace() []string {
	return e.stackTrace
}

func (e *AppErr) Unwrap() error {
	return e.wrapped
}

func (e *AppErr) LogValues(log logger.LogDict) {
	// Log error message.
	log.Str(errMsgKey, e.msg)

	// Log debug values.
	dict := log.NewDict()
	for key, value := range e.debugValues {
		dict.Str(key, value)
	}
	log.Dict(errDebugValuesKey, dict)

	// Log the stack trace (if any).
	if len(e.stackTrace) <= 0 {
		return
	}
	log.Strs(errStackKey, e.stackTrace...)
}

// Err creates new AppErr instance.
func Err(msg string, withStackTrace bool) *AppErr {
	err := AppErr{
		msg:         msg,
		debugValues: make(map[string]string),
	}
	if withStackTrace {
		err.stackTrace = getStackTrace()
	}
	return &err
}

func Wrap(err error, withStackTrace bool) *AppErr {
	// If the error provided is already *AppErr, it does not make sense to wrap it further.
	e, ok := err.(*AppErr)
	if ok {
		// Stack trace is required, but it was not generated beforehand. So try generating it now.
		if withStackTrace && len(e.stackTrace) == 0 {
			e.stackTrace = getStackTrace()
		}
		return e
	}

	e = Err(err.Error(), false)
	e.wrapped = err

	// Generate stack trace here to keep the skip level constant.
	if withStackTrace {
		e.stackTrace = getStackTrace()
	}

	return e
}

func getStackTrace() []string {
	rawStack := debug.Stack()
	if rawStack == nil {
		return []string{
			"{nil-stack-trace-returned-from-debug.Stack()}",
		}
	}

	stackStr := string(rawStack)
	split := strings.Split(stackStr, "\n")
	for i, s := range split {
		split[i] = strings.Replace(s, "\t", "    ", 1)
	}
	return split
}
