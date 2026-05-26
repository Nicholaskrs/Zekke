package helpers

import "fmt"

var _ error = (*AppErr)(nil)

type AppErr struct {
	msg     string
	wrapped error
}

func (e *AppErr) Error() string {
	if e.wrapped != nil {
		return fmt.Sprintf("%s: %s", e.msg, e.wrapped.Error())
	}
	return e.msg
}

func (e *AppErr) Unwrap() error {
	return e.wrapped
}

func Err(msg string) *AppErr {
	return &AppErr{msg: msg}
}

func Wrap(err error) *AppErr {
	return &AppErr{wrapped: err}
}
