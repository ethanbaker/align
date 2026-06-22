// Package main demonstrates how to wire the align core package together with
// the Discord and Telegram adapters to build a runnable scheduling bot.
//
// Steps:
//  1. Read secrets from the environment or a config file.
//  2. Construct one Adapter per platform (discord.New / telegram.New).
//  3. Choose a Persister (align.NopPersister for in-memory, or a custom
//     SQL-backed implementation for durability across restarts).
//  4. Create a Scheduler, passing the adapter registry and the persister.
//  5. Call scheduler.Start() to enable cron-driven scheduling, then block.
//
// OnContact and OnCompletion can also be called manually for testing or when
// a cron-driven flow is not required.
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/ethanbaker/align"
	"github.com/ethanbaker/align/discord"
	"github.com/ethanbaker/align/telegram"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	discordToken := os.Getenv("DISCORD_TOKEN")
	telegramToken := os.Getenv("TELEGRAM_TOKEN")

	// --- Discord session ---
	discordSession, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatalf("discord: %v", err)
	}
	if err = discordSession.Open(); err != nil {
		log.Fatalf("discord open: %v", err)
	}
	defer discordSession.Close()

	// --- Telegram session ---
	telegramBot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Fatalf("telegram: %v", err)
	}

	// --- Platform adapters (implement align.Contactor) ---
	discordAdapter := discord.New(discordSession)
	telegramAdapter := telegram.New(telegramBot)

	// --- Persister (in-memory; swap for a SQL implementation for durability) ---
	persister := align.NopPersister{}

	// --- Scheduler ---
	scheduler, err := align.NewScheduler(align.SchedulerOptions{
		Name:       "example",
		ConfigPath: "./config.yml",
		Contactors: map[string]align.Contactor{
			"discord":  discordAdapter,
			"telegram": telegramAdapter,
		},
		Persister: persister,
		OnContact: []func(){
			func() { log.Println("[example]: contact phase starting") },
		},
		OnCompletion: []func(days []align.Day){
			func(days []align.Day) {
				log.Printf("[example]: %d available day(s) found\n", len(days))
			},
		},
	})
	if err != nil {
		log.Fatalf("scheduler: %v", err)
	}

	// Start the cron-driven loop. OnContact and OnCompletion fire automatically
	// according to the cron strings in config.yml.
	if err = scheduler.Start(); err != nil {
		log.Fatalf("scheduler start: %v", err)
	}

	log.Println("align is running — press Ctrl-C to stop")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop
}
