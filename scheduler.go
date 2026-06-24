package align

import (
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

const dayDuration = int(time.Hour * 24)
const timeFormat = "Monday 01/02"

// SchedulerOptions is the single argument to NewScheduler. It carries all
// required configuration as well as optional callbacks.
type SchedulerOptions struct {
	// Name uniquely identifies this scheduler's persisted session.
	Name string

	// ConfigPath is the path to the YAML configuration file.
	ConfigPath string

	// Contactors maps each method name (matching Person.RequestMethod /
	// Person.ResponseMethod in the config) to its Contactor implementation.
	Contactors map[string]Contactor

	// Persister stores and restores Session state across restarts.
	// Use NopPersister for in-memory-only operation.
	Persister Persister

	// OnContact holds zero or more callbacks invoked at the start of the
	// contact phase, before any Contactor.Request calls are made.
	OnContact []func()

	// OnCompletion holds zero or more callbacks invoked after alignment is
	// computed, before result notifications are sent. Each callback receives
	// the days that meet the availability threshold.
	OnCompletion []func(days []Day)
}

// Scheduler orchestrates scheduling rounds. It holds a registry of Contactors
// (one per platform method name) and a Persister, and drives each session from
// initial contact through response gathering, alignment, and notification.
type Scheduler struct {
	contactors          map[string]Contactor
	persister           Persister
	config              *Config
	session             *Session
	loc                 *time.Location
	mu                  sync.Mutex
	contactListeners    []func()
	completionListeners []func(days []Day)
}

// NewScheduler reads the YAML config at opts.ConfigPath, restores any
// previously saved session via opts.Persister, and returns a ready Scheduler.
func NewScheduler(opts SchedulerOptions) (*Scheduler, error) {
	raw, err := os.ReadFile(opts.ConfigPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err = yaml.Unmarshal(raw, &config); err != nil {
		return nil, err
	}

	loc, err := time.LoadLocation(config.ContactTimezone)
	if err != nil {
		return nil, err
	}

	persister := opts.Persister
	if persister == nil {
		persister = NopPersister{}
	}

	s := &Scheduler{
		contactors:          opts.Contactors,
		persister:           persister,
		config:              &config,
		loc:                 loc,
		contactListeners:    opts.OnContact,
		completionListeners: opts.OnCompletion,
	}

	// Restore a previous session if one exists.
	saved, err := persister.LoadSession(opts.Name)
	if err != nil {
		return nil, err
	}

	if saved != nil {
		s.session = saved
	} else {
		s.session = &Session{
			Model: &gorm.Model{},
			Name:  opts.Name,
		}
	}

	// Give each Contactor the Persister and current Session so it can save
	// and restore its own in-flight entries, tagged with the Session's ID.
	for _, contactor := range s.contactors {
		contactor.SetPersister(persister, s.session)
	}

	if err := persister.SaveSession(s.session); err != nil {
		return nil, err
	}

	return s, nil
}

// Session returns a snapshot of the current session state.
func (s *Scheduler) Session() *Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.session
}

// Start registers the cron jobs for the contact and completion phases and
// begins the cron scheduler. It does not block.
func (s *Scheduler) Start() error {
	cronService := cron.New(cron.WithLocation(s.loc))

	if _, err := cronService.AddFunc(s.config.ContactTime, s.OnContact); err != nil {
		return err
	}
	if _, err := cronService.AddFunc(s.config.DeadlineTime, s.OnCompletion); err != nil {
		return err
	}

	cronService.Start()
	return nil
}

// OnContact runs the contact phase: it records the current time, then calls
// Contactor.Request for each person using their configured request method.
func (s *Scheduler) OnContact() {
	log.Println("[INFO]: starting contact phase")

	now := time.Now().In(s.loc)
	s.session.ContactDay = &now

	if err := s.persister.SaveSession(s.session); err != nil {
		log.Printf("[ERR]: error persisting session after contact: %v\n", err)
	}

	for _, fn := range s.contactListeners {
		fn()
	}

	dates := s.dates()
	for _, person := range s.config.Persons {
		contactor, ok := s.contactors[person.RequestMethod]
		if !ok {
			log.Printf("[ERR]: no contactor registered for method %q (person %q)\n", person.RequestMethod, person.Name)
			continue
		}
		if err := contactor.Request(person, dates); err != nil {
			log.Printf("[ERR]: Request failed for person %q: %v\n", person.Name, err)
		} else {
			log.Printf("[INFO]: Request sent to %q via %q\n", person.Name, person.RequestMethod)
		}
	}
}

// OnCompletion runs the completion phase: it gathers availability from each
// person, computes the intersection, invokes the OnCompletion callback, then
// notifies every person of the result.
func (s *Scheduler) OnCompletion() {
	log.Println("[INFO]: starting completion phase")

	availability := make(map[string]AvailabilityMap)
	dates := s.dates()

	// Gather availability for each person.
	for _, person := range s.config.Persons {
		contactor, ok := s.contactors[person.RequestMethod]
		if !ok {
			log.Printf("[ERR]: no contactor registered for method %q (person %q)\n", person.RequestMethod, person.Name)
			continue
		}

		a, err := contactor.Gather(person, dates)
		if err != nil {
			log.Printf("[ERR]: Gather failed for person %q: %v\n", person.Name, err)
			continue
		}

		s.mu.Lock()
		availability[person.Name] = a
		s.mu.Unlock()

		log.Printf("[INFO]: gathered availability for %q\n", person.Name)
	}

	// Identify persons who gave no availability at all.
	unknowns := []string{}
	for k, schedule := range availability {
		hasTrue := false
		for _, available := range schedule {
			if available {
				hasTrue = true
				break
			}
		}
		if !hasTrue || schedule == nil {
			delete(availability, k)
			unknowns = append(unknowns, k)
		}
	}

	for _, name := range unknowns {
		log.Printf("[INFO]: no availability from %q\n", name)
	}

	// Compute the best-effort intersection.
	var n int
	var days []Day
	for n = len(s.config.Persons) - len(unknowns); n > 0; n-- {
		days = align(availability, n)
		if len(days) > 0 {
			break
		}
	}

	for _, day := range days {
		log.Printf("[INFO]: available day: %v (%v)\n", day.Timestamp, strings.Join(day.AvailablePersons, ", "))
	}

	for _, fn := range s.completionListeners {
		fn(days)
	}

	// Notify each person of the result.
	for _, person := range s.config.Persons {
		contactor, ok := s.contactors[person.ResponseMethod]
		if !ok {
			log.Printf("[ERR]: no contactor registered for method %q (person %q)\n", person.ResponseMethod, person.Name)
			continue
		}
		if err := contactor.Notify(person, s.config.Title, days, unknowns, n); err != nil {
			log.Printf("[ERR]: Notify failed for person %q: %v\n", person.Name, err)
		} else {
			log.Printf("[INFO]: result sent to %q via %q\n", person.Name, person.ResponseMethod)
		}
	}

	// Void every Contactor's entries now that this round is finished, so
	// stale entries don't accumulate or get mistaken for the next round's.
	for name, contactor := range s.contactors {
		if err := contactor.VoidEntries(); err != nil {
			log.Printf("[ERR]: error voiding entries for contactor %q: %v\n", name, err)
		}
	}

	log.Println("[INFO]: completion phase done")
}

// dates returns the ordered list of date strings for the current session window,
// starting config.Offset days after ContactDay and spanning config.Interval days.
// It returns nil if no contact day has been set yet.
func (s *Scheduler) dates() []string {
	if s.session.ContactDay == nil {
		return nil
	}
	y, m, d := s.session.ContactDay.Date()
	base := time.Date(y, m, d, 0, 0, 0, 0, s.loc)

	dates := make([]string, 0, s.config.Interval)
	for i := s.config.Offset; i < s.config.Interval+s.config.Offset; i++ {
		dates = append(dates, base.Add(time.Duration(dayDuration*i)).Format(timeFormat))
	}
	return dates
}
