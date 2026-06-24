// Package main demonstrates how to use align with a durable SQL-backed
// Persister so that in-progress sessions survive a process restart.
//
// It uses the reusable align/persister/mysql package, which implements
// align.Persister with GORM + MySQL. See that package if you need to adapt
// it to another SQL driver that GORM supports (PostgreSQL, SQLite, etc.).
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
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/ethanbaker/align"
	"github.com/ethanbaker/align/adapter/discord"
	"github.com/ethanbaker/align/adapter/telegram"
	"github.com/ethanbaker/align/persister/mysql"
	mysql_driver "github.com/go-sql-driver/mysql"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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
	}

	persister, err := mysql.New(dsn.FormatDSN())
	if err != nil {
		log.Fatalf("persister: %v", err)
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config.yml"
	}

	// --- Scheduler ---
	scheduler, err := align.NewScheduler(align.SchedulerOptions{
		Name:       "example-sql",
		ConfigPath: configPath,
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
