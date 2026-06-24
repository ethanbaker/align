package align

// Persister is the extension point for durable storage. The Scheduler uses
// the Session methods to save/restore session state across restarts; the
// Entries methods are used directly by Contactors (via SetPersister) so each
// adapter can save and restore its own in-flight entries.
type Persister interface {
	// SaveSession writes the current session state to durable storage.
	SaveSession(session *Session) error

	// LoadSession retrieves a previously saved session by name.
	// It returns (nil, nil) when no session with that name exists.
	LoadSession(name string) (*Session, error)

	// SaveEntries writes entries under key to durable storage, on behalf of a
	// Contactor, tagged with session's ID. Entries with a zero ID are
	// inserted and have their generated ID written back; entries with a
	// non-zero ID are updated in place.
	SaveEntries(key string, session *Session, entries []Entry) error

	// LoadEntries retrieves the entries previously saved under key for
	// session. It returns (nil, nil) when no entries exist for that key.
	LoadEntries(key string, session *Session) ([]Entry, error)

	// VoidEntries discards every entry previously saved under key for
	// session, so a new round doesn't inherit stale entries from the last.
	VoidEntries(key string, session *Session) error
}

// NopPersister is a no-op Persister. Sessions and entries are kept only in
// memory.
type NopPersister struct{}

func (NopPersister) SaveSession(*Session) error                    { return nil }
func (NopPersister) LoadSession(string) (*Session, error)          { return nil, nil }
func (NopPersister) SaveEntries(string, *Session, []Entry) error   { return nil }
func (NopPersister) LoadEntries(string, *Session) ([]Entry, error) { return nil, nil }
func (NopPersister) VoidEntries(string, *Session) error            { return nil }
