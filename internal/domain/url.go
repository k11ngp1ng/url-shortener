package domain

import "time"

type URL struct {
	Code        string     `json:"code"`
	OriginalURL string     `json:"original_url"`
	Clicks      int64      `json:"clicks"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// Expired reports whether the link can no longer be used at the supplied time.
func (u URL) Expired(now time.Time) bool {
	return u.ExpiresAt != nil && !u.ExpiresAt.After(now)
}
