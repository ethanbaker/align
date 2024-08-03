package align_test

import (
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/ethanbaker/align"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

const path = "./config/test_config.yml"

func TestDiscord(t *testing.T) {
	require := require.New(t)

	// Read in discord credentials
	env, err := godotenv.Read("./config/modules/.env.discord")
	require.Nil(err)

	// Start a discordgo session
	session, err := discordgo.New("Bot " + env["DISCORD_TOKEN"])
	require.Nil(err)

	err = session.Open()
	require.Nil(err)

	// Create a new manager
	manager, err := align.CreateManager("test-discord", path, align.Options{
		UseSQL: false,
	})
	require.Nil(err)

	// Initialize the discord module
	align.InitDiscord(manager, session)

	// Perform the contact
	manager.OnContact()

	// Trap for gofunc in request
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Send response with on completion
	manager.OnCompletion()

	sc = make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
