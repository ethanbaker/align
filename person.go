package align

// Person represents a contactable person who provides feedback on what days they are free
type Person struct {
	Name           string `yaml:"name"`            // The person's name
	RequestMethod  string `yaml:"request_method"`  // The person's ideal contact method for availability requests
	ResponseMethod string `yaml:"response_method"` // The person's ideal contact method for responses
	ID             string `yaml:"id"`              // The person's ID used to contact them with a given method
}
