package enricher

import (
	"fmt"
	"time"
)

type ErrProviderRateLimit struct {
	ID        string
	Limit     int
	Remaining int
	ResetAt   time.Time
}

func (e *ErrProviderRateLimit) Error() string {
	return fmt.Sprintf("%s: rate limit exceeded. Limit: %d, Remaining: %d, Reset %s", e.ID, e.Limit, e.Remaining, e.ResetAt)
}

type ErrProviderAuth struct {
	ID       string
	Username string
}

func (e *ErrProviderAuth) Error() string {
	return fmt.Sprintf("%s: %s authentication failed", e.ID, e.Username)
}
