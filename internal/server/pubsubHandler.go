package server

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
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
		if len(cmd.Args) <= 0 {
			return resp.SimpleError("ERR wrong number of arguments for 'subscribe' command")
		}

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

	case "PUBLISH":
		if len(cmd.Args) != 2 {
			return resp.SimpleError("ERR wrong number of arguments for 'publish' command")
		}

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

	default:
		return resp.SimpleError("ERR unknown command")

	}
}
