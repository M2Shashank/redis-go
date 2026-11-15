package command

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func sortedSetHandler(cmd *resp.Command, dataStore *store.Store) resp.RESPType {
	switch cmd.Com {

	case "ZADD":
		if len(cmd.Args) < 3 || len(cmd.Args)%2 == 0 {
			return resp.SimpleError("ERR wrong number of arguments for 'zadd' command")
		}

		var ans int = 0
		for i := 1; i < len(cmd.Args); i += 2 {
			score, err := strconv.ParseFloat(cmd.Args[i], 64)
			if err != nil {
				return resp.SimpleError("ERR Invalid or Unsupported Arguments for ZADD command")
			}

			n, err := dataStore.ZADD(cmd.Args[0], score, cmd.Args[i+1])
			if err != nil {
				return resp.SimpleError("ERR " + err.Error())
			}
			ans = n
		}
		return resp.Integer(ans)

	default:
		return resp.SimpleError("ERR Invalid or Unsupported Command")

	}
}