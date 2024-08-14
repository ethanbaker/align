package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/ethanbaker/align"
)

func main() {
	// Create a new Discord session using the provided bot token.
	session, err := discordgo.New("Bot " + "YOUR_DISCORD_TOKEN")
	if err != nil {
		log.Fatal(err)
	}

	// Open a websocket connection to Discord and begin listening.
	err = session.Open()
	if err != nil {
		log.Fatal(err)
	}

	// Setup align
	manager, err := align.CreateManager("example-discord", "./config.yml", align.Options{
		UseSQL: false,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Initialize module
	align.InitDiscord(manager, session)

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Align is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	session.Close()
}
