package align

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
)

const DAY_DURATION = int(time.Hour * 24)
const TIME_FORMAT = "Monday 01/02"

// Manager struct represents a top level manager class
type Manager struct {
	Availability  map[string]map[string]bool // Persons' availabilities
	Config        *Config                    // Base align config
	ModuleConfigs map[string]interface{}     // configs for "modules"

	Loc  *time.Location // Timezone location for cron
	Stop bool           // When to stop requests from listening to users
	Edit *sync.Mutex    // Mutex for accessing manager fields

	Completed int // Number of completed subthreads
}

func (m *Manager) OnContact() {
	m.Stop = false
	m.Completed = 0

	for _, person := range m.Config.Persons {
		// Find the person's request method
		request, ok := requests[person.RequestMethod]
		if !ok {
			log.Printf("[ERR]: request method '%v' does not exist for person '%v'\n", person.ResponseMethod, person.Name)
		}

		// Perform the response
		if err := request(person, m); err != nil {
			m.Completed++
			log.Printf("[ERR]: error sending request (err: %v)\n", err)
		}
	}
}

func (m *Manager) OnCompletion() {
	// Stop the responses from listening
	m.Stop = true

	// Wait until all threads complete
	for m.Completed < len(m.Config.Persons) {
	}

	// Filter all false schedules (can never attend) and then check completion
	unknowns := []string{}
	for k, schedule := range m.Availability {
		allFalse := true
		for _, available := range schedule {
			allFalse = allFalse && !available

			if available {
				break
			}
		}

		if allFalse || schedule == nil {
			delete(m.Availability, k)
			unknowns = append(unknowns, k)
		}
	}

	// Everyone has sent in an availability schedule, so calculate available days
	var n int
	var days []Day
	for n = len(m.Config.Persons) - len(unknowns); n > 0; n-- {
		days = align(m.Availability, n)

		if len(days) > 0 {
			break
		}
	}

	// Send out available days to all persons
	for _, person := range m.Config.Persons {
		// Find the person's response method
		response, ok := responses[person.ResponseMethod]
		if !ok {
			log.Printf("[ERR]: response method '%v' does not exist for person '%v'\n", person.ResponseMethod, person.Name)
		}

		// Perform the response
		if err := response(person, m, days, unknowns, n); err != nil {
			log.Printf("[ERR]: error sending response (err: %v)\n", err)
		}
	}
}

// Generate a base availabiltiy map
func (m *Manager) GenerateAvailability() map[string]bool {
	availability := map[string]bool{}

	// Get the current date
	today := time.Now().In(m.Loc).Truncate(24 * time.Hour)

	for day := m.Config.Offset; day < m.Config.Interval+m.Config.Offset; day++ {
		// Generate the timestamp and set the availability to false
		timestamp := today.Add(time.Duration(DAY_DURATION * day)).Format(TIME_FORMAT)
		availability[timestamp] = false
	}

	return availability
}

// Create and initialize a new manager
func NewManager(path string) (*Manager, error) {
	manager := Manager{
		Availability:  make(map[string]map[string]bool),
		ModuleConfigs: make(map[string]interface{}),
		Config:        &Config{},
		Edit:          &sync.Mutex{},
		Completed:     0,
	}

	// Open the yaml config file
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Unmarshal the config
	if err = yaml.Unmarshal(file, manager.Config); err != nil {
		return nil, err
	}

	// Load the config timezone
	loc, err := time.LoadLocation(manager.Config.ContactTimezone)
	if err != nil {
		return nil, err
	}
	manager.Loc = loc

	// Start the cron service
	cronService := cron.New(cron.WithLocation(loc))
	cronService.Start()

	// Send availability requests according to the contact time cron string
	_, err = cronService.AddFunc(manager.Config.ContactTime, manager.OnContact)
	if err != nil {
		return nil, err
	}

	// Create a job that will align schedules at a given deadline
	_, err = cronService.AddFunc(manager.Config.DeadlineTime, manager.OnCompletion)
	if err != nil {
		return nil, err
	}

	return &manager, nil
}
