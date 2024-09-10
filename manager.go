package align

import (
	"log"
	"os"
	"strings"
	"sync"
	"time"

	mysql_driver "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Constant for a day's duration
const DAY_DURATION = int(time.Hour * 24)

// How the days will be formatted
const TIME_FORMAT = "Monday 01/02"

// Manager struct represents a top level manager class
type Manager struct {
	gorm.Model
	Name       string    `gorm:"uniqueIndex,length:256"` // The name identifier for the manager
	ContactDay time.Time // The day persons are contacted

	availability  map[string]map[string]bool `gorm:"-"` // Persons' availabilities
	config        *Config                    `gorm:"-"` // Base align config
	moduleConfigs map[string]interface{}     `gorm:"-"` // configs for "modules"
	loc           *time.Location             `gorm:"-"` // Timezone location for cron
	options       *Options                   `gorm:"-"` // Manager options

	edit *sync.Mutex `gorm:"-"` // Mutex for accessing manager fields
	db   *gorm.DB    `gorm:"-"` // Database for persistance of records
}

// OnContact contacts the persons listed in the config using their preferred method
func (m *Manager) OnContact() {
	log.Printf("[INFO]: starting contact\n")

	// Update the contact day
	m.ContactDay = time.Now().In(m.loc)
	if err := m.db.Save(m).Error; err != nil {
		log.Printf("[ERR]: error saving contact day to SQL, stopping (err: %v)\n", err)
		return
	}

	// For each person
	for _, person := range m.config.Persons {
		log.Printf("[INFO]: starting contact for '%v'\n", person.Name)

		// Find the person's request method
		request, ok := requests[person.RequestMethod]
		if !ok {
			log.Printf("[ERR]: request method '%v' does not exist for person '%v'\n", person.RequestMethod, person.Name)
		}

		// Perform the response
		if err := request(person, m); err != nil {
			log.Printf("[ERR]: error sending request (err: %v)\n", err)
		} else {
			log.Printf("[INFO]: request method '%v' completed for person '%v'\n", person.RequestMethod, person.Name)
		}
	}
}

// OnCompletion runs when the person deciding time completes
func (m *Manager) OnCompletion() {
	log.Println("[INFO]: starting completion")

	// Gather information from discord
	for _, person := range m.config.Persons {
		// Find the person's gather method
		gather, ok := gathers[person.RequestMethod]
		if !ok {
			log.Printf("[ERR]: gather method '%v' does not exist for person '%v'\n", person.RequestMethod, person.Name)
		}

		// Gather information for the person
		if err := gather(person, m); err != nil {
			log.Printf("[ERR]: error gathering response information (err: %v)\n", err)
		} else {
			log.Printf("[INFO]: gather method '%v' completed for person '%v'\n", person.RequestMethod, person.Name)
		}
	}

	log.Println("[INFO]: gathered information for all users")
	for name, schedule := range m.availability {
		log.Printf("[INFO]: availability for %v\n", name)
		for date, available := range schedule {
			log.Printf("[INFO]: - %v (%v)\n", date, available)
		}
	}

	// Filter schedules that are all false
	unknowns := []string{}
	for k, schedule := range m.availability {
		hasTrue := false
		for _, available := range schedule {
			hasTrue = hasTrue || available

			if hasTrue {
				break
			}
		}

		// Remove schedules that don't have at least one true entry or are nil
		if !hasTrue || schedule == nil {
			delete(m.availability, k)
			unknowns = append(unknowns, k)
		}
	}

	log.Println("[INFO]: found unknown users")
	for _, name := range unknowns {
		log.Printf("[INFO]: - %v\n", name)
	}

	// Everyone has sent in an availability schedule, so calculate available days
	var n int
	var days []day
	for n = len(m.config.Persons) - len(unknowns); n > 0; n-- {
		days = align(m.availability, n)

		if len(days) > 0 {
			break
		}
	}

	log.Println("[INFO]: calculated available days")
	for _, day := range days {
		log.Printf("[INFO]: - %v (with persons %v)\n", day.Timestamp, strings.Join(day.AvailablePersons, ", "))
	}

	// Send out available days to all persons
	for _, person := range m.config.Persons {
		// Find the person's response method
		response, ok := responses[person.ResponseMethod]
		if !ok {
			log.Printf("[ERR]: response method '%v' does not exist for person '%v'\n", person.ResponseMethod, person.Name)
		}

		// Perform the response
		if err := response(person, m, days, unknowns, n); err != nil {
			log.Printf("[ERR]: error sending response (err: %v)\n", err)
		} else {
			log.Printf("[INFO]: response method '%v' completed for person '%v'\n", person.RequestMethod, person.Name)
		}
	}

	log.Println("[INFO]: completion was successful")
}

// Generate a base availabiltiy map
func (m *Manager) generateAvailability() map[string]bool {
	availability := map[string]bool{}

	// Get the current date
	year, month, day := m.ContactDay.Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, m.loc)

	for day := m.config.Offset; day < m.config.Interval+m.config.Offset; day++ {
		// Generate the timestamp and set the availability to false
		timestamp := today.Add(time.Duration(DAY_DURATION * day)).Format(TIME_FORMAT)
		availability[timestamp] = false
	}

	return availability
}

// Create and initialize a new manager
func CreateManager(name string, path string, options Options) (*Manager, error) {
	var manager Manager
	manager.Name = name

	log.Println("[INFO]: reading yaml config file")

	// Open the yaml config file
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	log.Println("[INFO]: successfully read yaml config file")
	log.Println("[INFO]: unmarshalling yaml config file")

	// Unmarshal the config
	config := Config{}
	if err = yaml.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	log.Println("[INFO]: successfully unmarshalled yaml config file")

	if options.UseSQL {
		log.Println("[INFO]: setting up SQL")

		// Get the sql credentials if the manager is using SQL
		dsn := mysql_driver.Config{
			User:      config.Dsn.User,
			Passwd:    config.Dsn.Passwd,
			Net:       config.Dsn.Net,
			Addr:      config.Dsn.Addr,
			DBName:    config.Dsn.DBName,
			ParseTime: true,
		}

		log.Println("[INFO]: opening gorm database")

		// Open the gorm database
		db, err := gorm.Open(mysql.Open(dsn.FormatDSN()), &gorm.Config{})
		if err != nil {
			return nil, err
		}

		log.Println("[INFO]: migrating gorm databases")

		// Migrate the manager database
		if !db.Migrator().HasTable(&Manager{}) {
			if err = db.AutoMigrate(&Manager{}); err != nil {
				return nil, err
			}
		}

		log.Println("[INFO]: loading managers from SQL")

		// Load existing managers
		managers := []Manager{}
		if err := db.Model(&Manager{}).Find(&managers).Error; err != nil {
			return nil, err
		}

		log.Println("[INFO]: checking for existing managers")

		// Check for existing managers
		for _, m := range managers {
			if m.Name == name {
				log.Printf("[INFO]: returning existing manager with name %v", m.Name)

				manager = m
				manager.db = db
			}
		}

		// If there is no existing manager, save the current one
		if manager.db == nil {
			manager.db = db
			db.Save(&manager)
		}
	}

	// Populate manager fields
	manager.availability = make(map[string]map[string]bool)
	manager.moduleConfigs = make(map[string]interface{})
	manager.config = &config
	manager.edit = &sync.Mutex{}
	manager.options = &options

	log.Printf("[INFO]: loading timezone location '%v'\n", config.ContactTimezone)

	// Load the config timezone
	loc, err := time.LoadLocation(manager.config.ContactTimezone)
	if err != nil {
		return nil, err
	}
	manager.loc = loc

	log.Println("[INFO]: successfully loaded timezone")
	log.Println("[INFO]: starting cron service")

	// Start the cron service
	cronService := cron.New(cron.WithLocation(loc))
	cronService.Start()

	log.Println("[INFO]: adding 'ContactTime' cron func")

	// Send availability requests according to the contact time cron string
	_, err = cronService.AddFunc(manager.config.ContactTime, manager.OnContact)
	if err != nil {
		return nil, err
	}

	log.Println("[INFO]: adding 'OnCompletion' cron func")

	// Create a job that will align schedules at a given deadline
	_, err = cronService.AddFunc(manager.config.DeadlineTime, manager.OnCompletion)
	if err != nil {
		return nil, err
	}

	log.Println("[INFO]: returning newly created manager")

	return &manager, nil
}
