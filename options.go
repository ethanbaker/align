package align

// Options struct is used to provide custom options when creating a manager
type Options struct {
	// Whether or not align should use an SQL database to persist messages in case of power outages/etc
	UseSQL bool
}
