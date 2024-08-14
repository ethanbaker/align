package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/ethanbaker/align"
	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Start a telegram session
	telegramSession, err := telegram.NewBotAPI("YOUR_TELEGRAM_TOKEN")
	if err != nil {
		log.Fatal(err)
	}

	// Start a discordgo session
	discordSession, err := discordgo.New("Bot " + "YOUR_DISCORD_TOKEN")
	if err != nil {
		log.Fatal(err)
	}

	err = discordSession.Open()
	if err != nil {
		log.Fatal(err)
	}

	// Create a new manager
	manager, err := align.CreateManager("example-all", "./config.yml", align.Options{
		UseSQL: false,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the modules
	align.InitTelegram(manager, telegramSession)
	align.InitDiscord(manager, discordSession)

	// Perform the contact
	manager.OnContact()

	// Trap for gofunc in request
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Send response with on completion
	manager.OnCompletion()

	// Wait here until CTRL-C or other term signal is received.
	sc = make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down sessions
	discordSession.Close()
}
