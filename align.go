package align

// AvailabilityMap maps a date string to whether a person is available on that date.
type AvailabilityMap map[string]bool

// Day holds the result for a single date: which persons are free.
type Day struct {
	Timestamp        string   // The formatted date string
	AvailablePersons []string // Names of people available on this day
}

// align returns the days, in dates order, on which at least n people in s
// are available.
func align(s map[string]AvailabilityMap, dates []string, n int) []Day {
	days := make(map[string]*Day, len(dates))
	for _, date := range dates {
		days[date] = &Day{Timestamp: date, AvailablePersons: []string{}}
	}

	for name, availability := range s {
		for date, available := range availability {
			if available {
				if day, ok := days[date]; ok {
					day.AvailablePersons = append(day.AvailablePersons, name)
				}
			}
		}
	}

	filter := []Day{}
	for _, date := range dates {
		if day := days[date]; len(day.AvailablePersons) >= n {
			filter = append(filter, *day)
		}
	}

	return filter
}
