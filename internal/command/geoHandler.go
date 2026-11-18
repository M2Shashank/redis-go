package command

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func geoHandler(cmd *resp.Command, dataStore *store.Store) resp.RESPType {
	switch cmd.Com {
	case "GEOADD":
		if len(cmd.Args) < 4 || len(cmd.Args)%3 != 1 {
			return resp.SimpleError("ERR wrong number of arguments for 'geoadd' command")
		}

		newEntries := (len(cmd.Args) - 1) / 3
		return resp.Integer(newEntries)
	default:
		return resp.SimpleError("ERR Invalid or Unsupported Command")
	}
}
