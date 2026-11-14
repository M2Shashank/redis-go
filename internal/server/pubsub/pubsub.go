package pubsub

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

type ClientManager struct {
	// The channels the client is subscribed to.
	// Map for O(1) lookup
	subscribedChannels map[string]bool
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		subscribedChannels: make(map[string]bool),
	}
}

// client is in subscribed mode. Check if upcomming commands are allowed.
// Allowed Commands: (P)SUBSCRIBE, (P)UNSUBSCRIBE, PING, QUIT, RESET
func (cm *ClientManager) ValidateCommand(cmd *resp.Command) bool {
	switch cmd.Com {
	case "SUBSCRIBE", "PSUBSCRIBE", "UNSUBSCRIBE", "PUNSUBSCRIBE", "PING", "QUIT", "RESET":
		return true
	default:
		return false
	}
}

func (cm *ClientManager) HandleCommand(cmd *resp.Command) resp.RESPType {
	var result resp.RESPType

	switch cmd.Com {

	case "SUBSCRIBE":
		if len(cmd.Args) > 0 {
			cm.subscribedChannels[cmd.Args[0]] = true
		}

		result = resp.Array{}

		for _, ch := range cmd.Args {
			n := cm.SubscribeTo(ch)

			result = append(result.(resp.Array),
				resp.BulkString("subscribe"),
				resp.BulkString(ch),
				resp.Integer(n),
			)
		}

	default:
		result = resp.SimpleError("ERR unknown command")

	}

	return result
}
