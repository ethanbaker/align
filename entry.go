package align

// Entry is a generic, JSON-serializable record of platform-specific state
// that a Contactor must keep between Request and Gather (e.g. a Discord
// message ID or a Telegram poll ID). A Contactor saves and restores its own
// entries through the Persister given to it via SetPersister.
type Entry struct {
	// ID identifies this entry to a Persister across Save calls. It is zero
	// for an entry that has never been saved; a Persister that assigns its
	// own identifiers (e.g. an auto-increment row ID) should set it on first
	// Save so later Saves update the existing record instead of creating a
	// duplicate.
	ID uint `json:"id,omitempty"`

	// SessionID links this entry back to the Session it was created under.
	SessionID uint `json:"session_id,omitempty"`

	// Person is the name of the person this entry belongs to.
	Person string `json:"person"`

	// Index is the batch index this entry was created for (Request batches
	// dates into groups of up to seven; Index ties an entry back to its batch).
	Index int `json:"index"`

	// Fields holds adapter-specific identifiers (e.g. "messageID", "pollID",
	// "channelID", "chatID"), keyed by name.
	Fields map[string]string `json:"fields"`

	// Availability holds any availability accumulated for Person so far
	// outside of the Gather call itself (e.g. live Telegram poll votes
	// collected by a background listener). Adapters that compute availability
	// entirely within Gather can leave this nil.
	Availability AvailabilityMap `json:"availability,omitempty"`
}
