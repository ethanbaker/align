package align

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

var emojis = []string{
	"1️⃣",
	"2️⃣",
	"3️⃣",
	"4️⃣",
	"5️⃣",
	"6️⃣",
	"7️⃣",
}

const REQUEST_TEMPLATE_HEADER = `⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜

**Schedule for %v**`

const REQUEST_TEMPLATE_BODY = `%v
❌ - None

React with the corresponding emoji for dates you are free
`

const RESPONSE_TEMPLATE = `⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜

**Schedule results for %v**
%v

%v%v%v
⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜
`

// DiscordConfig holds all necessary fields for discord request/response functions to run successfully
type DiscordConfig struct {
	Session *discordgo.Session
}

// Initialize a discord config
func InitDiscord(manager *Manager, s *discordgo.Session) {
	manager.ModuleConfigs["discord"] = DiscordConfig{
		Session: s,
	}
}

// Request an availability schedule using discord credentials
func DiscordRequest(person Person, manager *Manager) error {
	// Attempt to load the discord config
	config, ok := manager.ModuleConfigs["discord"].(DiscordConfig)
	if !ok {
		return fmt.Errorf("discord config has not been initialized")
	}

	// Check if the session is valid
	if config.Session == nil {
		return fmt.Errorf("discord session is nil")
	}

	// Get a basic availability
	availability := manager.GenerateAvailability()

	// Generate dates
	allDates := []string{}

	today := time.Now().In(manager.Loc).Truncate(24 * time.Hour)
	for day := manager.Config.Offset; day < manager.Config.Interval+manager.Config.Offset; day++ {
		timestamp := today.Add(time.Duration(DAY_DURATION * day)).Format(TIME_FORMAT)
		allDates = append(allDates, timestamp)
	}

	// Create a private channel to DM the user
	channel, err := config.Session.UserChannelCreate(person.ID)
	if err != nil {
		return err
	}

	// Send the header message
	_, err = config.Session.ChannelMessageSend(channel.ID, fmt.Sprintf(REQUEST_TEMPLATE_HEADER, manager.Config.Title))
	if err != nil {
		return err
	}

	// Used to determine when the messages are done
	threads := 0
	completedThreads := 0

	// Send messages
	for i := 0; i < manager.Config.Interval; i += 7 {
		threads++

		dates := []string{}

		// Get a list of dates and the emoji - date paris for the message
		emojiDates := ""
		for j := 0; j < 7 && i+j < manager.Config.Interval; j++ {
			emojiDates += fmt.Sprintf("%v - %v\n", emojis[j], allDates[i+j])
			dates = append(dates, allDates[i+j])
		}

		// Send a DM
		m, err := config.Session.ChannelMessageSend(channel.ID, fmt.Sprintf(REQUEST_TEMPLATE_BODY, emojiDates))
		if err != nil {
			return err
		}

		// React to the DM with the emojis so the user can easily react
		for j := 0; j < len(emojis); j++ {
			if err = config.Session.MessageReactionAdd(channel.ID, m.ID, emojis[j]); err != nil {
				return err
			}
		}
		// Add a reaction for no date
		if err = config.Session.MessageReactionAdd(channel.ID, m.ID, "❌"); err != nil {
			return err
		}

		// Mutex for editing availability
		edit := sync.Mutex{}

		// Repeatedly wait for the user to react to this message
		go func() {
			for j := 0; !manager.Stop; j = (j + 1) % 7 {
				// Get message reactions for a given emoji
				users, err := config.Session.MessageReactions(channel.ID, m.ID, emojis[j], 2, "", "")
				if err != nil {
					log.Printf("[ERR]: error getting message reactions from user '%v' (err: %v)\n", person.Name, err)
				}

				// Set availability
				edit.Lock()
				availability[dates[j]] = len(users) == 2
				edit.Unlock()

				// If the user said no dates then reset all and break
				users, err = config.Session.MessageReactions(channel.ID, m.ID, "❌", 2, "", "")
				if err != nil {
					log.Printf("[ERR]: error getting message reactions from user '%v' (err: %v)\n", person.Name, err)
				}
				if len(users) == 2 {
					// Reset all availability
					for k := range availability {
						edit.Lock()
						availability[k] = false
						edit.Unlock()
					}
					break
				}
			}
			// On stop, update availability safely
			manager.Edit.Lock()

			manager.Availability[person.Name] = availability
			completedThreads++
			if completedThreads == threads {
				manager.Completed++
			}

			manager.Edit.Unlock()
		}()
	}

	return nil
}

func DiscordResponse(person Person, manager *Manager, days []Day, unknowns []string, available int) error {
	// Attempt to load the discord config
	config, ok := manager.ModuleConfigs["discord"].(DiscordConfig)
	if !ok {
		return fmt.Errorf("discord config has not been initialized")
	}

	// Check if the session is valid
	if config.Session == nil {
		return fmt.Errorf("discord session is nil")
	}

	// Calculate fraction
	fraction := fmt.Sprintf("%v/%v people available", available, len(manager.Config.Persons))

	// Concatenate days to a single string
	dayString := ""
	for _, day := range days {
		dayString += fmt.Sprintf("- %v (%v)\n", day.Timestamp, strings.Join(day.AvailablePersons, ", "))
	}

	// Concatenate unknowns into a single string
	unknownsString := ""
	for _, person := range unknowns {
		unknownsString += fmt.Sprintf("- %v\n", person)
	}

	unknownPrefix := ""
	if len(unknowns) > 0 {
		unknownPrefix = "No responses from:\n"
	}

	// Create a private channel to DM the user
	channel, err := config.Session.UserChannelCreate(person.ID)
	if err != nil {
		return err
	}

	// Send a message to the user
	_, err = config.Session.ChannelMessageSend(channel.ID, fmt.Sprintf(RESPONSE_TEMPLATE, manager.Config.Title, fraction, dayString, unknownPrefix, unknownsString))
	if err != nil {
		return err
	}

	return nil
}
