// Package main demonstrates how to use align with a durable SQL-backed
// Persister so that in-progress sessions survive a process restart.
//
// It implements align.Persister directly using GORM + MySQL. You can copy
// this implementation into your own project or adapt it to any other SQL
// driver that GORM supports (PostgreSQL, SQLite, etc.).
//
// Required environment variables:
//
//	DISCORD_TOKEN   — Discord bot token
//	TELEGRAM_TOKEN  — Telegram bot token
//	DB_USER         — MySQL username
//	DB_PASSWD       — MySQL password
//	DB_ADDR         — MySQL host:port  (e.g. 127.0.0.1:3306)
//	DB_NAME         — MySQL database name
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mysql_driver "github.com/go-sql-driver/mysql"
	"github.com/bwmarrin/discordgo"
	"github.com/ethanbaker/align"
	"github.com/ethanbaker/align/discord"
	"github.com/ethanbaker/align/telegram"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ----- SQL Persister -------------------------------------------------------

// sessionRow is the GORM model used to persist a single Session.
type sessionRow struct {
	gorm.Model
	Name         string     `gorm:"uniqueIndex;size:256"`
	ContactDay   *time.Time `gorm:"index"`
	Availability string     `gorm:"type:json"` // JSON-encoded map[string]AvailabilityMap
}

// SQLPersister implements align.Persister using a GORM database.
type SQLPersister struct {
	db *gorm.DB
}

// NewSQLPersister opens a GORM connection to the given DSN, migrates the
// sessions table, and returns a ready SQLPersister.
func NewSQLPersister(dsn string) (*SQLPersister, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("sql persister: open db: %w", err)
	}
	if err = db.AutoMigrate(&sessionRow{}); err != nil {
		return nil, fmt.Errorf("sql persister: migrate: %w", err)
	}
	return &SQLPersister{db: db}, nil
}

// Save upserts the session into the database.
func (p *SQLPersister) Save(session *align.Session) error {
	avJSON, err := json.Marshal(session.Availability)
	if err != nil {
		return fmt.Errorf("sql persister: marshal availability: %w", err)
	}

	row := sessionRow{
		Name:         session.Name,
		ContactDay:   session.ContactDay,
		Availability: string(avJSON),
	}

	return p.db.Where(sessionRow{Name: session.Name}).Assign(row).FirstOrCreate(&row).Error
}

// Load retrieves the session with the given name.
// It returns (nil, nil) when no session exists yet.
func (p *SQLPersister) Load(name string) (*align.Session, error) {
	var row sessionRow
	result := p.db.Where("name = ?", name).First(&row)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if result.Error != nil {
		return nil, fmt.Errorf("sql persister: load %q: %w", name, result.Error)
	}

	var availability map[string]align.AvailabilityMap
	if err := json.Unmarshal([]byte(row.Availability), &availability); err != nil {
		return nil, fmt.Errorf("sql persister: unmarshal availability for %q: %w", name, err)
	}

	return &align.Session{
		Name:         row.Name,
		ContactDay:   row.ContactDay,
		Availability: availability,
	}, nil
}

// ----- Consumer binary -----------------------------------------------------

func main() {
	// --- Platform sessions ---
	discordSession, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatalf("discord: %v", err)
	}
	if err = discordSession.Open(); err != nil {
		log.Fatalf("discord open: %v", err)
	}
	defer discordSession.Close()

	telegramBot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Fatalf("telegram: %v", err)
	}

	// --- SQL Persister ---
	dsn := mysql_driver.Config{
		User:      os.Getenv("DB_USER"),
		Passwd:    os.Getenv("DB_PASSWD"),
		Net:       "tcp",
		Addr:      os.Getenv("DB_ADDR"),
		DBName:    os.Getenv("DB_NAME"),
		ParseTime: true,
	}.FormatDSN()

	persister, err := NewSQLPersister(dsn)
	if err != nil {
		log.Fatalf("persister: %v", err)
	}

	// --- Scheduler ---
	scheduler, err := align.NewScheduler(align.SchedulerOptions{
		Name:       "example-sql",
		ConfigPath: "./config.yml",
		Contactors: map[string]align.Contactor{
			"discord":  discord.New(discordSession),
			"telegram": telegram.New(telegramBot),
		},
		Persister: persister,
		OnContact: []func(){
			func() { log.Println("[example-sql]: contact phase starting") },
		},
		OnCompletion: []func(days []align.Day){
			func(days []align.Day) {
				log.Printf("[example-sql]: %d available day(s) found\n", len(days))
			},
		},
	})
	if err != nil {
		log.Fatalf("scheduler: %v", err)
	}

	if err = scheduler.Start(); err != nil {
		log.Fatalf("scheduler start: %v", err)
	}

	log.Println("align (SQL) is running — press Ctrl-C to stop")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop
}
