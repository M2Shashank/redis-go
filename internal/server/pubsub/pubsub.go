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