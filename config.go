package align

// DSN represents sql credentials for the service
type DSN struct {
	User   string `yaml:"user"`
	Passwd string `yaml:"passwd"`
	Net    string `yaml:"tcp"`
	Addr   string `yaml:"addr"`
	DBName string `yaml:"dbname"`
}

// Person represents a contactable person who provides feedback on what days they are free
type Person struct {
	Name           string `yaml:"name"`            // The person's name
	RequestMethod  string `yaml:"request_method"`  // The person's ideal contact method for availability requests
	ResponseMethod string `yaml:"response_method"` // The person's ideal contact method for responses
	ID             string `yaml:"id"`              // The person's ID used to contact them with a given method
}

// Settings represent general configuration settings
type Settings struct {
	Title           string `yaml:"title"`         // The title of the group
	Interval        int    `yaml:"interval"`      // How many days to get availability for each cycle
	Offset          int    `yaml:"offset"`        // How many days after the contact date should availability gathering start
	ContactTimezone string `yaml:"timezone"`      // The timezone in which to contact persons
	ContactTime     string `yaml:"contact_time"`  // A cron string that shows when the persons should be contacted
	DeadlineTime    string `yaml:"deadline_time"` // A cron string that shows when the final decision should be made
}

// Config represents the configuration align will run off of
type Config struct {
	// Persons to run the application for
	Persons []Person `yaml:"persons"` // A list of persons to contact

	// SQL Credentials
	Dsn *DSN `yaml:"sql,omitempty"`

	// Application configuration settings
	Settings `yaml:"settings"`
}
