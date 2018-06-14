package rest

import (
	"fmt"
	"net/http"
)

type StatusCode int

func (code StatusCode) StatusCode() int {
	return int(code)
}

func (code StatusCode) Error() string {
	return http.StatusText(code.StatusCode())
}

func (code StatusCode) WithMessage(a ...interface{}) error {
	return &errStatus{
		code: code,
		msg:  fmt.Sprint(a...),
	}
}

func (code StatusCode) WithMessagef(format string, a ...interface{}) error {
	return &errStatus{
		code: code,
		msg:  fmt.Sprintf(format, a...),
	}
}

type errStatus struct {
	code StatusCode
	msg  string
}

func (e *errStatus) StatusCode() int {
	return int(e.code)
}

func (e *errStatus) Error() string {
	return e.msg
}

const (
	ErrBadRequest       StatusCode = http.StatusBadRequest
	ErrNotFound         StatusCode = http.StatusNotFound
	ErrConflict         StatusCode = http.StatusConflict
	ErrMethodNotAllowed StatusCode = http.StatusMethodNotAllowed
)
