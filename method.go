package align

// All possible request methods
var requests = map[string]func(Person, *Manager) error{
	"discord": DiscordRequest,
}

// All possible response methods
var responses = map[string]func(Person, *Manager, []Day, []string, int) error{
	"discord": DiscordResponse,
}
