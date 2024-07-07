package align

// Config represents the configuration align will run off of
type Config struct {
	// A list of persons to contact
	Persons []Person `yaml:"persons"`

	// The title of the group
	Title string `yaml:"title"`

	// How many days to get availability for each cycle
	Interval int `yaml:"interval"`

	// How many days after the contact date should availability gathering start
	Offset int `yaml:"offset"`

	// The timezone in which to contact persons
	ContactTimezone string `yaml:"contact_timezone"`

	// A cron string that shows when the persons should be contacted
	ContactTime string `yaml:"contact_time"`

	// A cron string that shows when the final decision should be made
	DeadlineTime string `yaml:"deadline_time"`
}
