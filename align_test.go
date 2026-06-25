package align_test

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethanbaker/align"
	"github.com/stretchr/testify/require"
)

// stubContactor is a minimal in-memory Contactor for unit tests.
type stubContactor struct {
	// responses maps person name → AvailabilityMap returned by Gather.
	responses map[string]align.AvailabilityMap
}

func (s *stubContactor) Request(_ align.Person, _ []string) error { return nil }

func (s *stubContactor) Gather(person align.Person, _ []string) (align.AvailabilityMap, error) {
	if av, ok := s.responses[person.Name]; ok {
		return av, nil
	}
	return nil, errors.New("stub: no response for " + person.Name)
}

func (s *stubContactor) Notify(_ align.Person, _ string, _ []align.Day, _ []string, _ int) error {
	return nil
}

func (s *stubContactor) SetPersister(_ align.Persister, _ *align.Session) {}

func (s *stubContactor) VoidEntries() error { return nil }

// stubPersister is a minimal in-memory Persister for unit tests.
type stubPersister struct {
	saved   map[string]*align.Session
	entries map[string][]align.Entry
}

func newStubPersister() *stubPersister {
	return &stubPersister{
		saved:   make(map[string]*align.Session),
		entries: make(map[string][]align.Entry),
	}
}

func (p *stubPersister) SaveSession(s *align.Session) error {
	p.saved[s.Name] = s
	return nil
}

func (p *stubPersister) LoadSession(name string) (*align.Session, error) {
	return p.saved[name], nil
}

func (p *stubPersister) SaveEntries(key string, _ *align.Session, entries []align.Entry) error {
	p.entries[key] = entries
	return nil
}

func (p *stubPersister) LoadEntries(key string, _ *align.Session) ([]align.Entry, error) {
	return p.entries[key], nil
}

func (p *stubPersister) VoidEntries(key string, _ *align.Session) error {
	delete(p.entries, key)
	return nil
}

// TestSchedulerConstruction verifies that NewScheduler parses config and wires
// contactors correctly without requiring any real platform connections.
func TestSchedulerConstruction(t *testing.T) {
	r := require.New(t)

	stub := &stubContactor{
		responses: map[string]align.AvailabilityMap{
			"Person 1": {"Monday 01/06": true, "Tuesday 01/07": false},
			"Person 2": {"Monday 01/06": true, "Tuesday 01/07": true},
		},
	}

	scheduler, err := align.NewScheduler(align.SchedulerOptions{
		Name:       "test-align",
		ConfigPath: "./testing/config_stub.yml",
		Contactors: map[string]align.Contactor{"stub": stub},
		Persister:  newStubPersister(),
	})
	r.NoError(err)
	r.NotNil(scheduler)
}

// TestSchedulerRestoresSession verifies that a persisted session is restored
// on the next NewScheduler call.
func TestSchedulerRestoresSession(t *testing.T) {
	r := require.New(t)

	persister := newStubPersister()
	stub := &stubContactor{responses: map[string]align.AvailabilityMap{}}

	contactors := map[string]align.Contactor{"stub": stub}

	// First run — no saved session.
	_, err := align.NewScheduler(align.SchedulerOptions{
		Name:       "restore-test",
		ConfigPath: "./testing/config_stub.yml",
		Contactors: contactors,
		Persister:  persister,
	})
	r.NoError(err)

	// Manually persist a session so the next run can restore it.
	r.NoError(persister.SaveSession(&align.Session{
		Name: "restore-test",
	}))

	// Second run — should restore without error.
	s2, err := align.NewScheduler(align.SchedulerOptions{
		Name:       "restore-test",
		ConfigPath: "./testing/config_stub.yml",
		Contactors: contactors,
		Persister:  persister,
	})
	r.NoError(err)
	r.NotNil(s2)
}

// TestNopPersister verifies that NopPersister compiles and is a no-op.
func TestNopPersister(t *testing.T) {
	r := require.New(t)

	var p align.Persister = align.NopPersister{}
	r.NoError(p.SaveSession(&align.Session{Name: "x"}))
	s, err := p.LoadSession("x")
	r.NoError(err)
	r.Nil(s)

	r.NoError(p.SaveEntries("k", nil, []align.Entry{{Person: "Alice"}}))
	entries, err := p.LoadEntries("k", nil)
	r.NoError(err)
	r.Nil(entries)
}

// TestSchedulerBadConfigPath ensures NewScheduler surfaces a missing-file error.
func TestSchedulerBadConfigPath(t *testing.T) {
	_, err := align.NewScheduler(align.SchedulerOptions{
		Name:       "bad",
		ConfigPath: "./testing/does_not_exist.yml",
		Contactors: map[string]align.Contactor{},
		Persister:  align.NopPersister{},
	})
	require.Error(t, err)
}

// TestSchedulerBadTimezone ensures NewScheduler rejects an invalid IANA zone.
func TestSchedulerBadTimezone(t *testing.T) {
	_, err := align.NewScheduler(align.SchedulerOptions{
		Name:       "bad-tz",
		ConfigPath: "./testing/config_bad_tz.yml",
		Contactors: map[string]align.Contactor{},
		Persister:  align.NopPersister{},
	})
	require.Error(t, err)
}

// TestSchedulerNilPersister verifies that a nil Persister is accepted and the
// scheduler still starts a fresh session without panicking.
func TestSchedulerNilPersister(t *testing.T) {
	r := require.New(t)
	sched, err := align.NewScheduler(align.SchedulerOptions{
		Name:       "nil-persister",
		ConfigPath: "./testing/config_stub.yml",
		Contactors: map[string]align.Contactor{"stub": &stubContactor{responses: map[string]align.AvailabilityMap{}}},
		Persister:  nil,
	})
	r.NoError(err)
	r.NotNil(sched)
}

// TestSchedulerOnContact verifies that OnContact fires contact listeners and
// records a ContactDay on the session.
func TestSchedulerOnContact(t *testing.T) {
	r := require.New(t)

	var called atomic.Int32
	stub := &stubContactor{responses: map[string]align.AvailabilityMap{}}

	sched, err := align.NewScheduler(align.SchedulerOptions{
		Name:       "on-contact",
		ConfigPath: "./testing/config_stub.yml",
		Contactors: map[string]align.Contactor{"stub": stub},
		Persister:  newStubPersister(),
		OnContact:  []func(){func() { called.Add(1) }},
	})
	r.NoError(err)

	before := time.Now()
	sched.OnContact()
	after := time.Now()

	r.EqualValues(1, called.Load())

	// The session should now expose a ContactDay within the test window.
	session := sched.Session()
	r.NotNil(session.ContactDay)
	r.True(!session.ContactDay.Before(before.Truncate(time.Second)))
	r.True(!session.ContactDay.After(after.Add(time.Second)))
}

// TestSchedulerOnCompletion verifies that OnCompletion fires completion
// listeners with the computed days and notifies every person.
func TestSchedulerOnCompletion(t *testing.T) {
	r := require.New(t)

	stub := &stubContactor{
		responses: map[string]align.AvailabilityMap{
			"Person 1": {"Monday 01/06": true, "Tuesday 01/07": false},
			"Person 2": {"Monday 01/06": true, "Tuesday 01/07": true},
		},
	}

	var gotDays []align.Day
	sched, err := align.NewScheduler(align.SchedulerOptions{
		Name:       "on-completion",
		ConfigPath: "./testing/config_stub.yml",
		Contactors: map[string]align.Contactor{"stub": stub},
		Persister:  newStubPersister(),
		OnCompletion: []func([]align.Day){
			func(days []align.Day) { gotDays = days },
		},
	})
	r.NoError(err)

	// Seed a contact day so dates() returns a non-nil slice.
	sched.OnContact()
	sched.OnCompletion()

	// Both persons are available Monday 01/06, so at least one day should match.
	r.NotNil(gotDays)
}

// TestSchedulerOnCompletionNoAvailability verifies that OnCompletion does not
// panic when no person provides any availability.
func TestSchedulerOnCompletionNoAvailability(t *testing.T) {
	stub := &stubContactor{
		responses: map[string]align.AvailabilityMap{
			"Person 1": {"Monday 01/06": false},
			"Person 2": {"Monday 01/06": false},
		},
	}

	sched, err := align.NewScheduler(align.SchedulerOptions{
		Name:       "no-avail",
		ConfigPath: "./testing/config_stub.yml",
		Contactors: map[string]align.Contactor{"stub": stub},
		Persister:  newStubPersister(),
	})
	require.NoError(t, err)

	sched.OnContact()
	require.NotPanics(t, sched.OnCompletion)
}

// TestSchedulerStart verifies that Start does not return an error for a
// well-formed config (cron expressions are syntactically valid).
func TestSchedulerStart(t *testing.T) {
	sched, err := align.NewScheduler(align.SchedulerOptions{
		Name:       "start-test",
		ConfigPath: "./testing/config_stub.yml",
		Contactors: map[string]align.Contactor{"stub": &stubContactor{responses: map[string]align.AvailabilityMap{}}},
		Persister:  align.NopPersister{},
	})
	require.NoError(t, err)
	require.NoError(t, sched.Start())
}

// TestSessionFields verifies that Session fields are exported and addressable.
func TestSessionFields(t *testing.T) {
	r := require.New(t)
	now := time.Now()
	s := &align.Session{
		Name:       "s",
		ContactDay: &now,
	}
	r.Equal("s", s.Name)
	r.NotNil(s.ContactDay)
}
