package align

// All possible request methods
var requests = map[string]func(Person, *Manager) error{
	"discord":  DiscordRequest,
	"telegram": TelegramRequest,
}

// All possible gather methods
var gathers = map[string]func(Person, *Manager) error{
	"discord":  DiscordGather,
	"telegram": TelegramGather,
}

// All possible response methods
var responses = map[string]func(Person, *Manager, []day, []string, int) error{
	"discord":  DiscordResponse,
	"telegram": TelegramResponse,
}
