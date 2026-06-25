package align

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestAlignEmpty returns nothing when the schedule map is empty.
func TestAlignEmpty(t *testing.T) {
	days := align(map[string]AvailabilityMap{}, nil, 1)
	require.Empty(t, days)
}

// TestAlignAllNil returns nothing when all maps are nil.
func TestAlignAllNil(t *testing.T) {
	days := align(map[string]AvailabilityMap{
		"Alice": nil,
		"Bob":   nil,
	}, []string{"Mon", "Tue"}, 1)
	require.Empty(t, days)
}

// TestAlignSinglePersonAvailable single person available on every date.
func TestAlignSinglePersonAvailable(t *testing.T) {
	r := require.New(t)
	days := align(map[string]AvailabilityMap{
		"Alice": {"Mon": true, "Tue": false},
	}, []string{"Mon", "Tue"}, 1)
	r.Len(days, 1)
	r.Equal("Mon", days[0].Timestamp)
	r.Contains(days[0].AvailablePersons, "Alice")
}

// TestAlignThresholdFiltering only returns days that meet the n threshold.
func TestAlignThresholdFiltering(t *testing.T) {
	r := require.New(t)
	schedules := map[string]AvailabilityMap{
		"Alice": {"Mon": true, "Tue": true, "Wed": false},
		"Bob":   {"Mon": true, "Tue": false, "Wed": false},
	}
	dates := []string{"Mon", "Tue", "Wed"}

	// Both available on Mon only.
	days := align(schedules, dates, 2)
	r.Len(days, 1)
	r.Equal("Mon", days[0].Timestamp)
	r.ElementsMatch([]string{"Alice", "Bob"}, days[0].AvailablePersons)

	// Alice available Mon and Tue; Bob only Mon.
	days = align(schedules, dates, 1)
	r.Len(days, 2)
}

// TestAlignThresholdTooHigh returns nothing when n exceeds total persons.
func TestAlignThresholdTooHigh(t *testing.T) {
	days := align(map[string]AvailabilityMap{
		"Alice": {"Mon": true},
	}, []string{"Mon"}, 5)
	require.Empty(t, days)
}

// TestDatesNilContactDay returns nil when no contact day is set.
func TestDatesNilContactDay(t *testing.T) {
	s := &Scheduler{
		config:  &Config{Settings: Settings{Interval: 7, Offset: 2}},
		session: &Session{},
		loc:     time.UTC,
	}
	require.Nil(t, s.dates())
}

// TestDatesCorrectRange verifies that dates are offset and span the right
// number of days in the configured format.
func TestDatesCorrectRange(t *testing.T) {
	r := require.New(t)

	contact := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // Monday
	s := &Scheduler{
		config:  &Config{Settings: Settings{Interval: 3, Offset: 1}},
		session: &Session{ContactDay: &contact},
		loc:     time.UTC,
	}

	dates := s.dates()
	r.Len(dates, 3)
	// offset=1 so first date is 2024-01-02 (Tuesday)
	r.Equal("Tuesday 01/02", dates[0])
	r.Equal("Wednesday 01/03", dates[1])
	r.Equal("Thursday 01/04", dates[2])
}

// TestDatesZeroOffset starts on the contact day itself.
func TestDatesZeroOffset(t *testing.T) {
	r := require.New(t)

	contact := time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC) // Monday
	s := &Scheduler{
		config:  &Config{Settings: Settings{Interval: 2, Offset: 0}},
		session: &Session{ContactDay: &contact},
		loc:     time.UTC,
	}

	dates := s.dates()
	r.Len(dates, 2)
	r.Equal("Monday 06/10", dates[0])
	r.Equal("Tuesday 06/11", dates[1])
}
