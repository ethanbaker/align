package align

// DSN holds SQL connection credentials.
type DSN struct {
	User   string `yaml:"user"`
	Passwd string `yaml:"passwd"`
	Net    string `yaml:"tcp"`
	Addr   string `yaml:"addr"`
	DBName string `yaml:"dbname"`
}

// Person is a contactable participant who provides their availability.
type Person struct {
	Name           string `yaml:"name"`            // Display name
	RequestMethod  string `yaml:"request_method"`  // Platform used to ask for availability
	ResponseMethod string `yaml:"response_method"` // Platform used to send the result
	ID             string `yaml:"id"`              // Platform-specific identifier
}

// Settings holds cron and window configuration for a scheduling round.
type Settings struct {
	Title           string `yaml:"title"`         // Group or event name shown in messages
	Interval        int    `yaml:"interval"`      // Number of days to poll per cycle
	Offset          int    `yaml:"offset"`        // Days after contact date to begin the window
	ContactTimezone string `yaml:"timezone"`      // IANA timezone for cron strings
	ContactTime     string `yaml:"contact_time"`  // Cron string: when to contact persons
	DeadlineTime    string `yaml:"deadline_time"` // Cron string: when to gather responses and send results
}

// Config is the top-level configuration loaded from YAML.
type Config struct {
	Persons  []Person `yaml:"persons"`
	Dsn      *DSN     `yaml:"sql,omitempty"`
	Settings `yaml:"settings"`
}
