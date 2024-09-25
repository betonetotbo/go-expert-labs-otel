package http_utils

import (
	"errors"
	"net/http"
)

type (
	HttpError struct {
		statusCode int
		msg        string
	}
)

func NewHttpError(statusCode int, msg string) *HttpError {
	return &HttpError{statusCode: statusCode, msg: msg}
}

func GetStatusCode(err error) int {
	var he *HttpError
	if errors.As(err, &he) {
		return he.statusCode
	}
	return http.StatusInternalServerError
}

func (e *HttpError) Error() string {
	return e.msg
}

func (e *HttpError) GetStatusCode() int {
	return e.statusCode
}
