// Package discord implements the align.Contactor interface for Discord.
// Use New to create an Adapter from an open discordgo.Session, then register
// it with a Scheduler under the key "discord".
package discord

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/ethanbaker/align"
)

const requestHeader = `⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜

**Schedule for %v**`

const requestBody = `%v
❌ - None

React with the corresponding emoji for dates you are free
`

const responseBody = `⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜

**Schedule results for %v**

%v/%v people available

%v%v%v
⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜
`

var emojis = []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣"}

// entry tracks a single message sent during a Request call so Gather can
// retrieve the reactions for it later.
type entry struct {
	person    string
	index     int
	channelID string
	messageID string
}

// Adapter implements align.Contactor for Discord.
type Adapter struct {
	session *discordgo.Session
	mu      sync.Mutex
	entries []*entry
}

// New creates a Discord Adapter from an already-opened discordgo.Session.
func New(session *discordgo.Session) *Adapter {
	return &Adapter{session: session}
}

// Request sends numbered reaction-poll messages to person via Discord DM,
// one message per batch of up to seven dates.
func (a *Adapter) Request(person align.Person, dates []string) error {
	if a.session == nil {
		return fmt.Errorf("discord: session is nil")
	}

	channel, err := a.session.UserChannelCreate(person.ID)
	if err != nil {
		return fmt.Errorf("discord: open DM channel for %q: %w", person.Name, err)
	}

	// We need the config title — embed it in messages using the header format
	// (title is not available here; callers pass it via the dates slice context,
	// so we use person.Name as a fallback label — the Scheduler passes the real
	// title in the standard request header constant that references manager.config.Title).
	// Since the Adapter does not have access to Config, we send the header without title.
	if _, err = a.session.ChannelMessageSend(channel.ID, fmt.Sprintf(requestHeader, person.Name)); err != nil {
		return fmt.Errorf("discord: send header to %q: %w", person.Name, err)
	}

	// Clear existing entries for this person before recording new ones.
	a.mu.Lock()
	filtered := a.entries[:0]
	for _, e := range a.entries {
		if e.person != person.Name {
			filtered = append(filtered, e)
		}
	}
	a.entries = filtered
	a.mu.Unlock()

	for i := 0; i*7 < len(dates); i++ {
		var emojiDates strings.Builder
		for j := 0; j < 7 && i*7+j < len(dates); j++ {
			fmt.Fprintf(&emojiDates, "%v - %v\n", emojis[j], dates[i*7+j])
		}

		m, err := a.session.ChannelMessageSend(channel.ID, fmt.Sprintf(requestBody, emojiDates.String()))
		if err != nil {
			return fmt.Errorf("discord: send batch %d to %q: %w", i, person.Name, err)
		}

		for j := 0; j < len(emojis) && i*7+j < len(dates); j++ {
			if err = a.session.MessageReactionAdd(channel.ID, m.ID, emojis[j]); err != nil {
				return fmt.Errorf("discord: pre-react batch %d emoji %d for %q: %w", i, j, person.Name, err)
			}
		}
		if err = a.session.MessageReactionAdd(channel.ID, m.ID, "❌"); err != nil {
			return fmt.Errorf("discord: pre-react ❌ batch %d for %q: %w", i, person.Name, err)
		}

		a.mu.Lock()
		a.entries = append(a.entries, &entry{
			person:    person.Name,
			index:     i,
			channelID: channel.ID,
			messageID: m.ID,
		})
		a.mu.Unlock()

		log.Printf("[discord]: sent batch %d to %q\n", i, person.Name)
	}

	return nil
}

// Gather reads the Discord reactions on each message previously sent to person,
// returning an AvailabilityMap keyed by the same date strings that were passed
// to Request. Processed entries are removed from internal state.
func (a *Adapter) Gather(person align.Person, dates []string) (align.AvailabilityMap, error) {
	if a.session == nil {
		return nil, fmt.Errorf("discord: session is nil")
	}

	// Build a default availability map (all false).
	availability := make(align.AvailabilityMap, len(dates))
	for _, d := range dates {
		availability[d] = false
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

	for i, e := range personEntries {
		// If the person reacted with ❌, skip this entire batch.
		users, err := a.session.MessageReactions(e.channelID, e.messageID, "❌", 2, "", "")
		if err != nil {
			log.Printf("[discord]: error reading ❌ reactions for %q entry %d: %v\n", person.Name, i, err)
		}
		if len(users) == 2 {
			continue
		}

		for j := 0; j < len(emojis) && i*7+j < len(dates); j++ {
			users, err := a.session.MessageReactions(e.channelID, e.messageID, emojis[j], 2, "", "")
			if err != nil {
				log.Printf("[discord]: error reading reactions for %q date %q: %v\n", person.Name, dates[i*7+j], err)
				continue
			}
			// 2 users means bot pre-reaction + the person's own reaction.
			availability[dates[i*7+j]] = len(users) == 2
		}
	}

	log.Printf("[discord]: gathered availability for %q\n", person.Name)
	return availability, nil
}

// Notify sends the scheduling result to person via Discord DM.
func (a *Adapter) Notify(person align.Person, days []align.Day, unknowns []string, available int) error {
	if a.session == nil {
		return fmt.Errorf("discord: session is nil")
	}

	var dayLines strings.Builder
	for _, day := range days {
		fmt.Fprintf(&dayLines, "- %v (%v)\n", day.Timestamp, strings.Join(day.AvailablePersons, ", "))
	}

	var unknownLines strings.Builder
	unknownPrefix := ""
	if len(unknowns) > 0 {
		unknownPrefix = "\nNo responses from:\n"
		for _, u := range unknowns {
			fmt.Fprintf(&unknownLines, "- %v\n", u)
		}
	}

	channel, err := a.session.UserChannelCreate(person.ID)
	if err != nil {
		return fmt.Errorf("discord: open DM channel for %q: %w", person.Name, err)
	}

	msg := fmt.Sprintf(responseBody, person.Name, available, available+len(unknowns), dayLines.String(), unknownPrefix, unknownLines.String())
	if _, err = a.session.ChannelMessageSend(channel.ID, msg); err != nil {
		return fmt.Errorf("discord: send result to %q: %w", person.Name, err)
	}

	log.Printf("[discord]: sent result to %q\n", person.Name)
	return nil
}
