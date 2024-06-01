package http

import (
	"errors"
	"net/http"
)

var (
	ErrorBadQueryParams = errors.New("bad query params")
)

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func buildApiError(code int, message string) apiError {
	return apiError{
		Code:    code,
		Message: message,
	}
}

func mapError(err error) apiError {
	switch {
	case errors.Is(err, ErrorBadQueryParams):
		return buildApiError(http.StatusBadRequest, "Invalid Query Params")
	default:
		return buildApiError(http.StatusInternalServerError, "Internal Server Error")
	}
}
