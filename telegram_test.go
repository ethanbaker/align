package align_test

import (
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/ethanbaker/align"
	telegram "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func TestTelegram(t *testing.T) {
	require := require.New(t)

	// Read in telegram credentials
	env, err := godotenv.Read("./config/telegram/.env.telegram")
	require.Nil(err)

	// Start a telegram session
	session, err := telegram.NewBotAPI(env["TELEGRAM_TOKEN"])
	require.Nil(err)
	session.Debug = true

	// Create a new manager
	manager, err := align.CreateManager("test-telegram", "./config/telegram/config.yml", align.Options{
		UseSQL: false,
	})
	require.Nil(err)

	// Initialize the telegram module
	align.InitTelegram(manager, session)

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

func TestTelegramPreSQL(t *testing.T) {
	require := require.New(t)

	// Read in telegram credentials
	env, err := godotenv.Read("./config/telegram/.env")
	require.Nil(err)

	// Start a telegram session
	session, err := telegram.NewBotAPI(env["TELEGRAM_TOKEN"])
	require.Nil(err)
	session.Debug = true

	// Create a new manager
	manager, err := align.CreateManager("test-telegram", "./config/telegram/config.yml", align.Options{
		UseSQL: true,
	})
	require.Nil(err)

	// Initialize the telegram module
	align.InitTelegram(manager, session)

	// Perform the contact
	manager.OnContact()

	// Assume power loss/program stops here
}

func TestTelegramPostSQL(t *testing.T) {
	require := require.New(t)

	// Read in telegram credentials
	env, err := godotenv.Read("./config/telegram/.env")
	require.Nil(err)

	// Start a telegram session
	session, err := telegram.NewBotAPI(env["TELEGRAM_TOKEN"])
	require.Nil(err)
	session.Debug = true

	// Create a new manager
	manager, err := align.CreateManager("test-telegram", "./config/telegram/config.yml", align.Options{
		UseSQL: true,
	})
	require.Nil(err)

	// Initialize the telegram module
	align.InitTelegram(manager, session)

	// Send response with on completion
	manager.OnCompletion()
}
