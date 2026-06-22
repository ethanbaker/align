package align_test

import (
	"errors"
	"testing"

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

func (s *stubContactor) Notify(_ align.Person, _ []align.Day, _ []string, _ int) error {
	return nil
}

// stubPersister is a minimal in-memory Persister for unit tests.
type stubPersister struct {
	saved map[string]*align.Session
}

func newStubPersister() *stubPersister {
	return &stubPersister{saved: make(map[string]*align.Session)}
}

func (p *stubPersister) Save(s *align.Session) error {
	p.saved[s.Name] = s
	return nil
}

func (p *stubPersister) Load(name string) (*align.Session, error) {
	return p.saved[name], nil
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
	r.NoError(persister.Save(&align.Session{
		Name:         "restore-test",
		Availability: map[string]align.AvailabilityMap{},
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
	r.NoError(p.Save(&align.Session{Name: "x", Availability: map[string]align.AvailabilityMap{}}))
	s, err := p.Load("x")
	r.NoError(err)
	r.Nil(s)
}
