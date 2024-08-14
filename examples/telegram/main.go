package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethanbaker/align"
	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Start a telegram session
	session, err := telegram.NewBotAPI("YOUR_TELEGRAM_TOKEN")
	if err != nil {
		log.Fatal(err)
	}

	// Setup align
	manager, err := align.CreateManager("example-discord", "./config.yml", align.Options{
		UseSQL: false,
	})
	if err != nil {
		log.Println("error setting up manager,", err)
	}

	// Initialize module
	align.InitTelegram(manager, session)

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Align is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
