// Package telegram implements the align.Contactor interface for Telegram.
// Use New to create an Adapter from a telegram-bot-api BotAPI, then register
// it with a Scheduler under the key "telegram".
//
// Telegram limitations to be aware of:
//   - Bots cannot initiate conversations; each person must have previously sent
//     the bot a message (e.g. /start) so the bot knows their chat ID.
//   - Telegram servers retain poll updates for only 24 hours, so the bot must
//     process updates at least once per day to capture all responses.
package telegram

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/ethanbaker/align"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const requestHeader = "Schedule\n\nPlease select the dates you are free"

const responseBody = `Schedule results for %v

%v/%v people available

%v%v%v`

// entry tracks a single Telegram poll sent during a Request call.
type entry struct {
	person    string
	index     int
	pollID    string
	messageID int
	chatID    int64
}

// Adapter implements align.Contactor for Telegram.
type Adapter struct {
	bot          *tgbotapi.BotAPI
	mu           sync.Mutex
	entries      []*entry
	availability map[string]align.AvailabilityMap // person name → date → available
}

// New creates a Telegram Adapter and starts a background goroutine to process
// poll updates from Telegram.
func New(bot *tgbotapi.BotAPI) *Adapter {
	a := &Adapter{
		bot:          bot,
		availability: make(map[string]align.AvailabilityMap),
	}
	go a.listenUpdates()
	return a
}

// listenUpdates processes incoming Telegram updates, updating availability
// whenever a poll answer is received.
func (a *Adapter) listenUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := a.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Poll == nil {
			continue
		}
		poll := update.Poll

		a.mu.Lock()
		var matched *entry
		for _, e := range a.entries {
			if e.pollID == poll.ID {
				matched = e
				break
			}
		}

		if matched == nil {
			a.mu.Unlock()
			log.Printf("[telegram]: received update for unknown poll %q\n", poll.ID)
			continue
		}

		av, ok := a.availability[matched.person]
		if !ok {
			a.mu.Unlock()
			log.Printf("[telegram]: no availability map for person %q\n", matched.person)
			continue
		}

		for _, opt := range poll.Options {
			av[opt.Text] = opt.VoterCount > 0
			log.Printf("[telegram]: %q availability on %q: %v\n", matched.person, opt.Text, opt.VoterCount > 0)
		}
		a.mu.Unlock()
	}
}

// Request sends Telegram polls to person, one per batch of up to seven dates.
func (a *Adapter) Request(person align.Person, dates []string) error {
	if a.bot == nil {
		return fmt.Errorf("telegram: bot is nil")
	}

	userID, err := strconv.ParseInt(person.ID, 10, 64)
	if err != nil {
		return fmt.Errorf("telegram: invalid user ID %q for %q: %w", person.ID, person.Name, err)
	}

	// Initialise a fresh availability map for this person.
	av := make(align.AvailabilityMap, len(dates))
	for _, d := range dates {
		av[d] = false
	}

	a.mu.Lock()
	// Clear previous entries for this person.
	filtered := a.entries[:0]
	for _, e := range a.entries {
		if e.person != person.Name {
			filtered = append(filtered, e)
		}
	}
	a.entries = filtered
	a.availability[person.Name] = av
	a.mu.Unlock()

	for i := 0; i*7 < len(dates); i++ {
		options := []string{}
		for j := 0; j < 7 && i*7+j < len(dates); j++ {
			options = append(options, dates[i*7+j])
		}

		poll := tgbotapi.NewPoll(userID, requestHeader, options...)
		poll.AllowsMultipleAnswers = true

		m, err := a.bot.Send(poll)
		if err != nil {
			return fmt.Errorf("telegram: send poll batch %d to %q: %w", i, person.Name, err)
		}

		a.mu.Lock()
		a.entries = append(a.entries, &entry{
			person:    person.Name,
			index:     i,
			pollID:    m.Poll.ID,
			messageID: m.MessageID,
			chatID:    userID,
		})
		a.mu.Unlock()

		log.Printf("[telegram]: sent poll batch %d to %q\n", i, person.Name)
	}

	return nil
}

// Gather stops the open polls for person and returns the availability
// that was accumulated by the background update listener.
func (a *Adapter) Gather(person align.Person, dates []string) (align.AvailabilityMap, error) {
	if a.bot == nil {
		return nil, fmt.Errorf("telegram: bot is nil")
	}

	userID, err := strconv.ParseInt(person.ID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("telegram: invalid user ID %q for %q: %w", person.ID, person.Name, err)
	}

	// Extract and remove entries for this person.
	a.mu.Lock()
	var personEntries []*entry
	remaining := a.entries[:0]
	for _, e := range a.entries {
		if e.person == person.Name {
			personEntries = append(personEntries, e)
		} else {
			remaining = append(remaining, e)
		}
	}
	a.entries = remaining
	a.mu.Unlock()

	sort.Slice(personEntries, func(i, j int) bool {
		return personEntries[i].index < personEntries[j].index
	})

	for _, e := range personEntries {
		if _, err := a.bot.StopPoll(tgbotapi.NewStopPoll(userID, e.messageID)); err != nil {
			log.Printf("[telegram]: error stopping poll %q for %q: %v\n", e.pollID, person.Name, err)
		}
	}

	a.mu.Lock()
	av, ok := a.availability[person.Name]
	a.mu.Unlock()

	if !ok {
		// Return an all-false map rather than an error so the scheduler can
		// treat this person as "no response" instead of a hard failure.
		av = make(align.AvailabilityMap, len(dates))
		for _, d := range dates {
			av[d] = false
		}
	}

	log.Printf("[telegram]: gathered availability for %q\n", person.Name)
	return av, nil
}

// Notify sends the scheduling result to person as a Telegram message.
func (a *Adapter) Notify(person align.Person, days []align.Day, unknowns []string, available int) error {
	if a.bot == nil {
		return fmt.Errorf("telegram: bot is nil")
	}

	userID, err := strconv.ParseInt(person.ID, 10, 64)
	if err != nil {
		return fmt.Errorf("telegram: invalid user ID %q for %q: %w", person.ID, person.Name, err)
	}

	dayLines := ""
	for _, day := range days {
		dayLines += fmt.Sprintf("- %v (%v)\n", day.Timestamp, strings.Join(day.AvailablePersons, ", "))
	}

	unknownLines := ""
	unknownPrefix := ""
	if len(unknowns) > 0 {
		unknownPrefix = "\nNo responses from:\n"
		for _, u := range unknowns {
			unknownLines += fmt.Sprintf("- %v\n", u)
		}
	}

	text := fmt.Sprintf(responseBody, person.Name, available, available+len(unknowns), dayLines, unknownPrefix, unknownLines)
	msg := tgbotapi.NewMessage(userID, text)
	if _, err = a.bot.Send(msg); err != nil {
		return fmt.Errorf("telegram: send result to %q: %w", person.Name, err)
	}

	log.Printf("[telegram]: sent result to %q\n", person.Name)
	return nil
}
