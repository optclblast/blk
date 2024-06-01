package getblock

import "errors"

var (
	// ErrorRateLimitExceeded is thrown when
	// the number of requests has exceeded the allowed limit
	ErrorRateLimitExceeded = errors.New("api rate limit exceeded")
)
