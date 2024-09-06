package align

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

/* ---- TYPES ---- */

// DiscordConfig holds all necessary fields for discord request/response functions to run successfully
type DiscordConfig struct {
	Session *discordgo.Session
}

type discordEntry struct {
	gorm.Model

	Person    string // The person's name this entry is related to
	Index     int    // The index of this entry
	ChannelID string // The discord channel ID this entry represents
	MessageID string // The discord message ID this entry represents

	// The manager this entry is related to
	Manager   *Manager
	ManagerID *int
}

/* ---- GLOBALS ---- */

var discordEntries []*discordEntry

var emojis = []string{
	"1️⃣",
	"2️⃣",
	"3️⃣",
	"4️⃣",
	"5️⃣",
	"6️⃣",
	"7️⃣",
}

const discordRequestHeader = `⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜

**Schedule for %v**`

const discordRequestBody = `%v
❌ - None

React with the corresponding emoji for dates you are free
`

const discordResponseBody = `⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜

**Schedule results for %v**

%v/%v people available

%v%v%v
⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜⬜
`

/* ---- FUNCTIONS ---- */

// Initialize a discord config
func InitDiscord(manager *Manager, s *discordgo.Session) {
	log.Println("[INFO]: initializing discord config")

	manager.moduleConfigs["discord"] = DiscordConfig{
		Session: s,
	}

	if !manager.options.UseSQL {
		return
	}

	// If using SQL, populate discord entries
	if !manager.db.Migrator().HasTable(&discordEntry{}) {
		if err := manager.db.AutoMigrate(&discordEntry{}); err != nil {
			log.Fatalf("[ERR]: cannot migrate discordEntry object (err: %v)\n", err)
		}
	}

	if err := manager.db.Model(&discordEntry{}).Where("id = ?", fmt.Sprint(manager.ID)).Find(&discordEntries).Error; err != nil {
		log.Fatalf("[ERR]: cannot read discord entries from database (err: %v)\n", err)
	}
}

// Request an availability schedule using discord
func DiscordRequest(person Person, manager *Manager) error {
	log.Println("[INFO]: loading discord config")

	// Attempt to load the discord config
	config, ok := manager.moduleConfigs["discord"].(DiscordConfig)
	if !ok {
		return fmt.Errorf("discord config has not been initialized")
	}

	// Check if the session is valid
	if config.Session == nil {
		return fmt.Errorf("discord session is nil")
	}

	log.Println("[INFO]: generating availability dates")

	// Generate all dates in the availability map
	year, month, day := time.Now().In(manager.loc).Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, manager.loc)

	dates := []string{}
	for day := manager.config.Offset; day < manager.config.Interval+manager.config.Offset; day++ {
		timestamp := today.Add(time.Duration(DAY_DURATION * day)).Format(TIME_FORMAT)
		dates = append(dates, timestamp)
	}

	log.Printf("[INFO]: opening discord channel to id '%v'\n", person.ID)

	// Create a private channel to DM the user
	channel, err := config.Session.UserChannelCreate(person.ID)
	if err != nil {
		return err
	}

	log.Println("[INFO]: sending discord header")

	// Send the header message
	_, err = config.Session.ChannelMessageSend(channel.ID, fmt.Sprintf(discordRequestHeader, manager.config.Title))
	if err != nil {
		return err
	}

	log.Println("[INFO]: sending discord messages")

	// Send messages
	for i := 0; i*7 < manager.config.Interval; i++ {
		// Get a list of dates and the emoji - date paris for the message
		emojiDates := ""
		for j := 0; j < 7 && i*7+j < manager.config.Interval; j++ {
			emojiDates += fmt.Sprintf("%v - %v\n", emojis[j], dates[i*7+j])
		}

		// Send a DM
		m, err := config.Session.ChannelMessageSend(channel.ID, fmt.Sprintf(discordRequestBody, emojiDates))
		if err != nil {
			return err
		}

		// React to the DM with the emojis so the user can easily react
		for j := 0; j < len(emojis) && i*7+j < manager.config.Interval; j++ {
			if err = config.Session.MessageReactionAdd(channel.ID, m.ID, emojis[j]); err != nil {
				return err
			}
		}
		// Add a reaction for no date
		if err = config.Session.MessageReactionAdd(channel.ID, m.ID, "❌"); err != nil {
			return err
		}

		// Add this message as a recorded entry
		entry := discordEntry{
			Person:    person.Name,
			Index:     i,
			ChannelID: channel.ID,
			MessageID: m.ID,
			Manager:   manager,
		}
		discordEntries = append(discordEntries, &entry)

		// If using SQL, add to SQL database
		if manager.options.UseSQL {
			log.Println("[INFO]: adding discord entry to SQL")
			if err := manager.db.Save(&entry).Error; err != nil {
				log.Printf("[ERR]: error saving discord entry to SQL (err: %v)\n", err)
			}
		}
	}

	return nil
}

// Read a response for availability using discord
func DiscordGather(person Person, manager *Manager) error {
	log.Println("[INFO]: loading discord config")

	// Attempt to load the discord config
	config, ok := manager.moduleConfigs["discord"].(DiscordConfig)
	if !ok {
		return fmt.Errorf("discord config has not been initialized")
	}

	// Check if the session is valid
	if config.Session == nil {
		return fmt.Errorf("discord session is nil")
	}

	log.Println("[INFO]: generating availability dates")

	// Generate all dates in the availability map
	year, month, day := time.Now().In(manager.loc).Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, manager.loc)

	dates := []string{}
	for day := manager.config.Offset; day < manager.config.Interval+manager.config.Offset; day++ {
		timestamp := today.Add(time.Duration(DAY_DURATION * day)).Format(TIME_FORMAT)
		dates = append(dates, timestamp)
	}

	// Generate an availability for the person
	availability := manager.generateAvailability()

	log.Printf("[INFO]: collecting discord entries for '%v'", person.Name)

	// Filter for entries for this specific person
	var entries []*discordEntry
	for i := 0; i < len(discordEntries); i++ {
		// On matching entry, add to local array and remove from global
		if discordEntries[i].Person == person.Name {
			entries = append(entries, discordEntries[i])

			// If using SQL, remove from SQL database
			if manager.options.UseSQL {
				if err := manager.db.Delete(&discordEntries[i]).Error; err != nil {
					log.Printf("[ERR]: error deleting discord entry from SQL (err: %v)\n", err)
				}
			}

			discordEntries = append(discordEntries[:i], discordEntries[i+1:]...)
			i--
		}
	}

	// Sort entries based on index
	for i := 1; i < len(entries); i++ {
		cur := entries[i]
		j := i - 1

		for j >= 0 && entries[j].Index > cur.Index {
			entries[j+1] = entries[j]
			j--
		}
		entries[j+1] = cur
	}

	for i, entry := range entries {
		log.Printf("[INFO]: determining reactions for '%v' with entry number '%v' and message id '%v'\n", entry.Person, entry.Index, entry.MessageID)

		// Check if the user responded with an X
		users, err := config.Session.MessageReactions(entry.ChannelID, entry.MessageID, "❌", 2, "", "")
		if err != nil {
			log.Printf("[ERR]: error getting message reactions from user '%v' (err: %v)\n", person.Name, err)
		}

		// If the user responded with an X, skip this entry (they're not available)
		if len(users) == 2 {
			continue
		}

		// Check for reactions to individual dates
		for j := 0; j < len(emojis) && i*7+j < manager.config.Interval; j++ {
			// Get message reactions for a given emoji
			users, err := config.Session.MessageReactions(entry.ChannelID, entry.MessageID, emojis[j], 2, "", "")
			if err != nil {
				log.Printf("[ERR]: error getting message reactions from user '%v' (err: %v)\n", person.Name, err)
			}

			// Set availability based on the reaction count
			availability[dates[i*7+j]] = len(users) == 2
		}
	}

	// Log the user's availability
	for date, status := range availability {
		log.Printf("[INFO]: user '%v' availability status on %v is %v\n", person.Name, date, status)
	}

	// Update the user's availability in the manager
	manager.edit.Lock()
	manager.availability[person.Name] = availability
	manager.edit.Unlock()

	return nil
}

// Send a user a response summary on discord
func DiscordResponse(person Person, manager *Manager, days []day, unknowns []string, available int) error {
	log.Println("[INFO]: loading discord config")

	// Attempt to load the discord config
	config, ok := manager.moduleConfigs["discord"].(DiscordConfig)
	if !ok {
		return fmt.Errorf("discord config has not been initialized")
	}

	// Check if the session is valid
	if config.Session == nil {
		return fmt.Errorf("discord session is nil")
	}

	log.Println("[INFO]: building response string")

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
		unknownPrefix = "\nNo responses from:\n"
	}

	// Create a private channel to DM the user
	channel, err := config.Session.UserChannelCreate(person.ID)
	if err != nil {
		return err
	}

	// Format the message to be sent
	str := fmt.Sprintf(discordResponseBody,
		manager.config.Title,
		available,
		len(manager.config.Persons),
		dayString,
		unknownPrefix,
		unknownsString,
	)

	log.Printf("[INFO]: sending response message\n%v\n", str)

	// Send a message to the user
	_, err = config.Session.ChannelMessageSend(channel.ID, str)
	if err != nil {
		return err
	}

	return nil
}
