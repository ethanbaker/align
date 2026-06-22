package align

import "time"

// Session holds the in-progress state of a single scheduling round.
// All fields are exported and JSON-friendly so that a Persister can
// serialize and restore the session across restarts.
type Session struct {
	Name         string                     `json:"name"`
	ContactDay   *time.Time                 `json:"contact_day,omitempty"`
	Availability map[string]AvailabilityMap `json:"availability"`
}
