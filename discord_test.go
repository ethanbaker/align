package align_test

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/ethanbaker/align"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func TestDiscord(t *testing.T) {
	require := require.New(t)

	// Read in discord credentials
	env, err := godotenv.Read("./config/discord/.env")
	require.Nil(err)

	// Start a discordgo session
	session, err := discordgo.New("Bot " + env["DISCORD_TOKEN"])
	require.Nil(err)

	err = session.Open()
	require.Nil(err)

	// Create a new manager
	manager, err := align.CreateManager("test-discord", "./config/discord/config.yml", align.Options{
		UseSQL: false,
	})
	require.Nil(err)

	// Initialize the discord module
	align.InitDiscord(manager, session)

	// Perform the contact
	manager.OnContact()

	// Trap for gofunc in request
	log.Printf("[TEST]: waiting for interrupt\n")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Send response with on completion
	manager.OnCompletion()

	sc = make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

// Test normal progression of the application with SQL enabled (program doesn't stop)
func TestDiscordSQL(t *testing.T) {
	require := require.New(t)

	// Read in discord credentials
	env, err := godotenv.Read("./config/discord/.env")
	require.Nil(err)

	// Start a discordgo session
	session, err := discordgo.New("Bot " + env["DISCORD_TOKEN"])
	require.Nil(err)

	err = session.Open()
	require.Nil(err)

	// Create a new manager
	manager, err := align.CreateManager("test-discord", "./config/discord/config.yml", align.Options{
		UseSQL: true,
	})
	require.Nil(err)

	// Initialize the discord module
	align.InitDiscord(manager, session)

	// Perform the contact
	manager.OnContact()

	// Trap for gofunc in request
	/*
		log.Printf("[TEST]: waiting for interrupt\n")
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		<-sc
	*/

	// Send response with on completion
	manager.OnCompletion()
}

func TestDiscordPreSQL(t *testing.T) {
	require := require.New(t)

	// Read in discord credentials
	env, err := godotenv.Read("./config/discord/.env")
	require.Nil(err)

	// Start a discordgo session
	session, err := discordgo.New("Bot " + env["DISCORD_TOKEN"])
	require.Nil(err)

	err = session.Open()
	require.Nil(err)

	// Create a new manager
	manager, err := align.CreateManager("test-discord", "./config/discord/config.yml", align.Options{
		UseSQL: true,
	})
	require.Nil(err)

	// Initialize the discord module
	align.InitDiscord(manager, session)

	// Perform the contact
	manager.OnContact()

	// Assume power loss/program stops here
}

func TestDiscordPostSQL(t *testing.T) {
	require := require.New(t)

	// Read in discord credentials
	env, err := godotenv.Read("./config/discord/.env")
	require.Nil(err)

	// Start a discordgo session
	session, err := discordgo.New("Bot " + env["DISCORD_TOKEN"])
	require.Nil(err)

	err = session.Open()
	require.Nil(err)

	// Create a new manager
	manager, err := align.CreateManager("test-discord", "./config/discord/config.yml", align.Options{
		UseSQL: true,
	})
	require.Nil(err)

	// Initialize the discord module
	align.InitDiscord(manager, session)

	// Send response with on completion
	manager.OnCompletion()
}

// Test multiple progressions of the application with SQL enabled (program doesn't stop)
func TestDiscordSQLMultiple(t *testing.T) {
	require := require.New(t)

	// Read in discord credentials
	env, err := godotenv.Read("./config/discord/.env")
	require.Nil(err)

	// Start a discordgo session
	session, err := discordgo.New("Bot " + env["DISCORD_TOKEN"])
	require.Nil(err)

	err = session.Open()
	require.Nil(err)

	// Create a new manager
	manager, err := align.CreateManager("test-discord", "./config/discord/config.yml", align.Options{
		UseSQL: true,
	})
	require.Nil(err)

	// Initialize the discord module
	align.InitDiscord(manager, session)

	for i := 0; i < 3; i++ {
		log.Printf("[TEST]: testing iteration %v\n", i)

		// Perform the contact
		manager.OnContact()

		// Trap for gofunc in request
		log.Printf("[TEST]: waiting for interrupt\n")
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		<-sc

		// Initialize the discord module
		align.InitDiscord(manager, session)

		// Send response with on completion
		manager.OnCompletion()
	}
}
