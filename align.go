package align

// AvailabilityMap maps a date string to whether a person is available on that date.
type AvailabilityMap map[string]bool

// Day holds the result for a single date: which persons are free.
type Day struct {
	Timestamp        string   // The formatted date string
	AvailablePersons []string // Names of people available on this day
}

// align returns the days on which at least n people in s are available.
func align(s map[string]AvailabilityMap, n int) []Day {
	schedules := make(map[string]AvailabilityMap)
	for k, v := range s {
		if v != nil {
			schedules[k] = v
		}
	}

	if len(schedules) == 0 {
		return []Day{}
	}

	days := map[string]Day{}
	for key1 := range schedules {
		for date := range schedules[key1] {
			days[date] = Day{Timestamp: date, AvailablePersons: []string{}}
		}
		break
	}

	for name, availability := range schedules {
		for date, available := range availability {
			if available {
				if day, ok := days[date]; ok {
					day.AvailablePersons = append(day.AvailablePersons, name)
					days[date] = day
				}
			}
		}
	}

	filter := []Day{}
	for _, day := range days {
		if len(day.AvailablePersons) >= n {
			filter = append(filter, day)
		}
	}

	return filter
}
