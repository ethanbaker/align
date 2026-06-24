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

	// Notify delivers the final scheduling result to person. title is the
	// scheduling round's display name (Config.Settings.Title), shown in the
	// notification in place of person's name.
	Notify(person Person, title string, days []Day, unknowns []string, available int) error

	// VoidEntries discards the Contactor's currently persisted entries (e.g.
	// open Discord messages or Telegram polls left over from a previous
	// round), both in memory and in its Persister, if one has been set. The
	// Scheduler calls this at the start of each contact phase so stale
	// entries don't accumulate or get mistaken for the new round's.
	VoidEntries() error

	// SetPersister gives the Contactor a Persister to use for saving and
	// loading its own in-flight entries (e.g. open Discord messages or
	// Telegram polls), according to its own spec, along with the current
	// Session so saved entries can be tagged with its ID. Implementations
	// should load any previously saved entries immediately, and save their
	// current entries to p whenever they change.
	SetPersister(p Persister, session *Session)
}
