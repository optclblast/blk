package getblock

import "errors"

var (
	ErrorRateLimitExceeded = errors.New("api rate limit exceeded")
)
