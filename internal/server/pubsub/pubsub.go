package pubsub

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

type Manager struct {
	// Clients are subscribers. client id's to channel mapping.
	Subscribers  map[string]map[string]bool
	IsSubscribed map[string]bool // true if client is in subscribed mode (after SUBSCRIBE command). false otherwise.
	// Client id to Bool mapping.

	// Publishers are channels. Channel to client id's mapping.
	Publishers map[string]map[string]bool
}

func NewManager() *Manager {
	return &Manager{
		Subscribers:  make(map[string]map[string]bool),
		IsSubscribed: make(map[string]bool),
		Publishers:   make(map[string]map[string]bool),
	}
}

func (m *Manager) HandleCommand(cmd *resp.Command, clientId string) resp.RESPType {
	switch cmd.Com {

	case "PING":
		var result resp.RESPType
		if m.IsSubscribed[clientId] {
			result = resp.Array{
				resp.BulkString("pong"),
				resp.BulkString(""),
			}
		} else {
			result = resp.SimpleString("PONG")
		}
		return result

	case "SUBSCRIBE":
		if len(cmd.Args) <= 0 {
			return resp.SimpleError("ERR wrong number of arguments for 'subscribe' command")
		}

		result := resp.Array{}
		for _, ch := range cmd.Args {
			n := m.SubscribeTo(ch, clientId)

			result = append(result,
				resp.BulkString("subscribe"),
				resp.BulkString(ch),
				resp.Integer(n),
			)
		}
		m.IsSubscribed[clientId] = true
		return result

	case "PUBLISH":
		if len(cmd.Args) != 2 {
			return resp.SimpleError("ERR wrong number of arguments for 'publish' command")
		}

		n := m.PublishMessage(cmd.Args[0], cmd.Args[1])
		return resp.Integer(n)

	default:
		return resp.SimpleError("ERR unknown command")

	}
}
