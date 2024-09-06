package align

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

/** ---- TYPES ---- */

// TelegramConfig holds all necessary fields for discord request/response functions to run successfully
type TelegramConfig struct {
	Session *telegram.BotAPI
	Updates *telegram.UpdatesChannel
}

type telegramEntry struct {
	gorm.Model

	Person    string // The person's name this entry is related to
	Index     int    // The index of this entry
	PollID    string // The telegram poll ID to get results from
	MessageID int    // The telegram message ID to get results from

	// The manager this entry is related to
	Manager   *Manager
	ManagerID *int
}

/* ---- GLOBALS ---- */

var telegramEntries []*telegramEntry

const telegramRequestHeader = `**Schedule for %v**

Please enter the dates you are free`

const telegramResponseBody = `**Schedule results for %v**

%v/%v people available

%v%v%v`

/* ---- FUNCTIONS ---- */

// Initialize a telegram config
func InitTelegram(manager *Manager, s *telegram.BotAPI) {
	log.Println("[INFO]: initializing telegram config")

	// Add the config
	manager.moduleConfigs["telegram"] = TelegramConfig{
		Session: s,
	}

	if manager.options.UseSQL {
		// If using SQL, populate telegram entries
		if !manager.db.Migrator().HasTable(&telegramEntry{}) {
			if err := manager.db.AutoMigrate(&telegramEntry{}); err != nil {
				log.Fatalf("[ERR]: cannot migrate telegramEntry object (err: %v)\n", err)
			}
		}

		if err := manager.db.Model(&telegramEntry{}).Find(&telegramEntries).Error; err != nil {
			log.Fatalf("[ERR]: cannot read telegram entries from database (err: %v)\n", err)
		}

		// Generate a template availability for each person in the entries
		for _, entry := range telegramEntries {
			_, ok := manager.availability[entry.Person]
			if !ok {
				manager.availability[entry.Person] = manager.generateAvailability()
			}
		}
	}

	// Create an update channel and listen for updates
	u := telegram.NewUpdate(0)
	u.Timeout = 60
	updates := s.GetUpdatesChan(u)

	go func() {
		// Process incoming updates
		for update := range updates {
			// Discard any message that isn't a poll
			if update.Poll == nil {
				continue
			}
			poll := update.Poll

			// Find the availability of a person who updated a poll
			var availability map[string]bool
			var person Person
			for _, entry := range telegramEntries {
				if entry.PollID == poll.ID {
					// Get the availability of the person
					a, ok := manager.availability[entry.Person]
					if !ok {
						log.Printf("[WARN]: cannot get availability from person '%v'\n", entry.Person)
						break
					}

					availability = a

					// Find the associated person from the entry
					for _, p := range manager.config.Persons {
						if p.Name == entry.Person {
							person = p
							break
						}
					}
					break
				}
			}

			// If no availability is found, continue
			if availability == nil {
				continue
			}

			// Update the person's availability based on the poll results
			for _, option := range poll.Options {
				available := option.VoterCount > 0
				availability[option.Text] = available

				log.Printf("[INFO]: availability for '%v' on '%v' is %v\n", person.Name, option.Text, available)
			}
		}
	}()

}

// Request an availability schedule using telegram
func TelegramRequest(person Person, manager *Manager) error {
	log.Println("[INFO]: loading telegram config")

	// Attempt to load the telegram config
	config, ok := manager.moduleConfigs["telegram"].(TelegramConfig)
	if !ok {
		return fmt.Errorf("telegram config has not been initialized")
	}

	// Check if the session is valid
	if config.Session == nil {
		return fmt.Errorf("telegram session is nil")
	}

	// Generate an availability for the person
	availability := manager.generateAvailability()

	manager.edit.Lock()
	manager.availability[person.Name] = availability
	manager.edit.Unlock()

	log.Println("[INFO]: generating availability dates")

	// Generate all dates in the availability map
	year, month, day := time.Now().In(manager.loc).Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, manager.loc)

	dates := []string{}
	for day := manager.config.Offset; day < manager.config.Interval+manager.config.Offset; day++ {
		timestamp := today.Add(time.Duration(DAY_DURATION * day)).Format(TIME_FORMAT)
		dates = append(dates, timestamp)
	}

	log.Println("[INFO]: formatting user ID")

	// Format user ID
	userID, err := strconv.Atoi(person.ID)
	if err != nil {
		return err
	}

	// Generate the header
	header := fmt.Sprintf(telegramRequestHeader, manager.config.Title)

	log.Println("[INFO]: sending telegram messages")

	// Send messages
	for i := 0; i*7 < manager.config.Interval; i++ {
		// Get the dates to send
		options := []string{}
		for j := 0; j < 7 && i*7+j < manager.config.Interval; j++ {
			options = append(options, dates[i*7+j])
		}

		// Create a telegram poll
		poll := telegram.NewPoll(int64(userID), header, options...)
		poll.AllowsMultipleAnswers = true

		// Send the poll
		m, err := config.Session.Send(poll)
		if err != nil {
			return err
		}

		// Add this message as a recorded entry
		entry := telegramEntry{
			Person:    person.Name,
			Index:     i,
			PollID:    m.Poll.ID,
			MessageID: int(m.Chat.ID),
			Manager:   manager,
		}
		telegramEntries = append(telegramEntries, &entry)

		// If using SQL, add to SQL database
		if manager.options.UseSQL {
			log.Println("[INFO]: adding discord entry to SQL")
			if err := manager.db.Save(&entry).Error; err != nil {
				log.Printf("[ERR]: error saving telegram entry to SQL (err: %v)\n", err)
			}
		}
	}

	return nil
}

// Read a response for availability using telegram
func TelegramGather(person Person, manager *Manager) error {
	log.Println("[INFO]: loading telegram config")

	// Attempt to load the telegram config
	config, ok := manager.moduleConfigs["telegram"].(TelegramConfig)
	if !ok {
		return fmt.Errorf("telegram config has not been initialized")
	}

	// Check if the session is valid
	if config.Session == nil {
		return fmt.Errorf("telegram session is nil")
	}

	log.Println("[INFO]: formatting user ID")

	// Format user ID
	userID, err := strconv.Atoi(person.ID)
	if err != nil {
		return err
	}

	log.Printf("[INFO]: collecting telegram entries for '%v'", person.Name)

	// Filter for entries for this specific person
	var entries []*telegramEntry
	for i := 0; i < len(telegramEntries); i++ {
		// On matching entry, add to local array and remove from global
		if telegramEntries[i].Person == person.Name {
			entries = append(entries, telegramEntries[i])

			// If using SQL, remove from SQL database
			if manager.options.UseSQL {
				if err := manager.db.Delete(&telegramEntries[i]).Error; err != nil {
					log.Printf("[ERR]: error deleting telegram entry from SQL (err: %v)\n", err)
				}
			}

			telegramEntries = append(telegramEntries[:i], telegramEntries[i+1:]...)
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

	log.Printf("[INFO]: stopping telegram polls for '%v'\n", person.Name)

	for _, entry := range entries {
		// Stop the poll represented by this entry
		_, err := config.Session.StopPoll(telegram.NewStopPoll(int64(userID), entry.MessageID))
		if err != nil {
			return err
		}
	}

	// Get user's availability
	availability, ok := manager.availability[person.Name]
	if !ok {
		return fmt.Errorf("cannot find availability for '%v'", person.Name)
	}

	// Log the user's availability
	for date, status := range availability {
		log.Printf("[INFO]: user '%v' availability status on %v is %v\n", person.Name, date, status)
	}

	return nil
}

// Send a user a response summary on telegram
func TelegramResponse(person Person, manager *Manager, days []day, unknowns []string, available int) error {
	log.Println("[INFO]: loading telegram config")

	// Attempt to load the telegram config
	config, ok := manager.moduleConfigs["telegram"].(TelegramConfig)
	if !ok {
		return fmt.Errorf("telegram config has not been initialized")
	}

	// Check if the session is valid
	if config.Session == nil {
		return fmt.Errorf("telegram session is nil")
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

	// Format user ID
	userID, err := strconv.Atoi(person.ID)
	if err != nil {
		return err
	}

	// Format the message to be send
	str := fmt.Sprintf(telegramResponseBody,
		manager.config.Title,
		available,
		len(manager.config.Persons),
		dayString,
		unknownPrefix,
		unknownsString,
	)

	log.Printf("[INFO]: sending response message\n%v\n", str)

	// Send a message to the user
	_, err = config.Session.Send(telegram.NewMessage(int64(userID), str))
	if err != nil {
		return err
	}

	return nil
}
