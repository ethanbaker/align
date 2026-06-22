package align

// Persister is the extension point for durable session storage.
// The Scheduler calls Save after every state change and Load at startup
// so an in-progress session survives a process restart.
type Persister interface {
	// Save writes the current session state to durable storage.
	Save(session *Session) error

	// Load retrieves a previously saved session by name.
	// It returns (nil, nil) when no session with that name exists.
	Load(name string) (*Session, error)
}

// NopPersister is a no-op Persister. Sessions are kept only in memory.
type NopPersister struct{}

func (NopPersister) Save(*Session) error          { return nil }
func (NopPersister) Load(string) (*Session, error) { return nil, nil }
