package align

// All possible request methods
var requests = map[string]func(Person, *Manager) error{
	"discord": DiscordRequest,
}

// All possible gather methods
var gathers = map[string]func(Person, *Manager) error{
	"discord": DiscordGather,
}

// All possible response methods
var responses = map[string]func(Person, *Manager, []day, []string, int) error{
	"discord": DiscordResponse,
}
