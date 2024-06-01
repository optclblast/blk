package http

import (
	"errors"
	"net/http"

	"github.com/optclblast/blk/internal/infrastructure/getblock"
)

var (
	// ErrorBadQueryParams is thrown wheh query parameters are invalid
	ErrorBadQueryParams = errors.New("bad query params")
)

// api error dto object
type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// buildApiError returns a new apiError instance built from error code and error message
func buildApiError(
	code int,
	message string,
) apiError {
	return apiError{
		Code:    code,
		Message: message,
	}
}

// mapError maps internal errors to its API representation
func mapError(err error) apiError {
	switch {
	case errors.Is(err, ErrorBadQueryParams):
		return buildApiError(http.StatusBadRequest, "Invalid Query Params")
	case errors.Is(err, getblock.ErrorRateLimitExceeded):
		return buildApiError(
			http.StatusTooManyRequests,
			"GetBlock API rate limit exceeded! Type again later",
		)
	default:
		return buildApiError(http.StatusInternalServerError, "Internal Server Error")
	}
}
