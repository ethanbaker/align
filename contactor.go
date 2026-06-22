package align

// Contactor is the extension point for platform adapters.
// The Scheduler calls these methods to drive a scheduling session.
// Implementations must be safe for concurrent use.
type Contactor interface {
	// Request sends an availability poll to person for the given date strings.
	Request(person Person, dates []string) error

	// Gather collects person's poll responses and returns their availability.
	// It is called once per person after the deadline, using the same dates
	// slice that was passed to the corresponding Request call.
	Gather(person Person, dates []string) (AvailabilityMap, error)

	// Notify delivers the final scheduling result to person.
	Notify(person Person, days []Day, unknowns []string, available int) error
}
