package server

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"

	Error "github.com/codecrafters-io/redis-starter-go/internal/errors"
)

func (s *Server) PubSubHandler(cmd *resp.Command, clientId string) resp.RESPType {
	switch cmd.Com {
	case "PING":
		var result resp.RESPType
		if s.Pubsub.InSubsciberMode[clientId] {
			result = resp.Array{
				resp.BulkString("pong"),
				resp.BulkString(""),
			}
		} else {
			result = resp.SimpleString("PONG")
		}
		return result

	case "SUBSCRIBE":
		result := resp.Array{}
		for _, ch := range cmd.Args {
			n := s.Pubsub.SubscribeTo(ch, clientId)

			result = append(result,
				resp.BulkString("subscribe"),
				resp.BulkString(ch),
				resp.Integer(n),
			)
		}
		s.Pubsub.InSubsciberMode[clientId] = true
		return result

	case "UNSUBSCRIBE":
		result := resp.Array{}
		for _, ch := range cmd.Args {
			n := s.Pubsub.UnsubscribeFrom(ch, clientId)

			result = append(result,
				resp.BulkString("unsubscribe"),
				resp.BulkString(ch),
				resp.Integer(n),
			)
		}
		return result

	case "PUBLISH":
		mssg := resp.Array{
			resp.BulkString("message"),
			resp.BulkString(cmd.Args[0]),
			resp.BulkString(cmd.Args[1]),
		}
		response := resp.Encode(mssg)

		n, list := s.Pubsub.GetChannelSubscribers(cmd.Args[0])
		for _, clientID := range list {
			var client *Client = s.Clients[clientID]
			client.Send(response)
		}

		return resp.Integer(n)
	}

	return resp.SimpleError(Error.ERR_INVALID.Error())
}
