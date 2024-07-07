package align

type Day struct {
	Timestamp        string   // The timestamp of the day
	AvailablePersons []string // Available people
	// ...
}

// align a bunch of schedules together, returning a list of days n people are free
func align(s map[string]map[string]bool, n int) []Day {
	// Make a copy of the schedule map without nil availabilities
	schedules := make(map[string]map[string]bool)
	for k, v := range s {
		if v != nil {
			schedules[k] = v
		}
	}

	// Make sure schedules are present
	if len(schedules) == 0 {
		return []Day{}
	}

	// Initialize the list of days
	days := map[string]Day{}
	for key1 := range schedules { // We only need one schedule, so just grab the first one and break after
		for date := range schedules[key1] {
			days[date] = Day{
				Timestamp:        date,
				AvailablePersons: []string{},
			}
		}
		break
	}

	// For each person, check if they are available. If they are, add them to the day's count
	for name, availability := range schedules {
		for date, available := range availability {
			// If the person is available, add them to the available day
			if available {
				if day, ok := days[date]; ok {
					day.AvailablePersons = append(day.AvailablePersons, name)
					days[date] = day
				}
			}
		}
	}

	// Filter out for days that meet the 'n' criteria
	filter := []Day{}
	for _, day := range days {
		if len(day.AvailablePersons) >= n {
			filter = append(filter, day)
		}
	}

	return filter
}
