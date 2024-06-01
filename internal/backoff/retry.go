package backoff

import (
	"errors"
	"fmt"
	"time"

	"github.com/optclblast/blk/internal/infrastructure/getblock"
)

// DeciderFunc decides whether the function should be retried.
type DeciderFunc func(err error) bool

// WithRetry reties fn if error was approved by DeciderFunc and attempts
// and there are still attempts remaining.
func WithRetry(n int, fn func() error, decider DeciderFunc) error {
	// At least 1 attempt
	if n <= 0 {
		n = 1
	}

	var err error

	for i := 0; i < n; i++ {
		err = fn()

		if err != nil && !decider(err) {
			return fmt.Errorf("decided not to repeat. %w", err)
		}
	}

	return err
}

// As a getblock free plan users, we have only 60 rps and
// RateLimiterDecider is a small workaround
func RateLimiterDecider() DeciderFunc {
	return func(err error) bool {
		if errors.Is(err, getblock.ErrorRateLimitExceeded) {
			<-time.After(time.Second)
			return true
		}

		return false
	}
}
