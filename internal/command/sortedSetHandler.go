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
				return resp.SimpleError("ERR value is not a valid float")
			}

			done, err := dataStore.ZADD(cmd.Args[0], score, cmd.Args[i+1])
			if err != nil {
				return resp.SimpleError(err.Error())
			}

			if done {
				ans ++
			}
		}
		return resp.Integer(ans)

	case "ZRANK":
		if len(cmd.Args) != 2 {
			return resp.SimpleError("ERR wrong number of arguments for 'zrank' command")
		}

		index, err := dataStore.ZRANK(cmd.Args[0], cmd.Args[1])
		if err != nil {
			return resp.Null{}
		}

		return resp.Integer(index)

	case "ZRANGE":
		if len(cmd.Args) < 3 {
			return resp.SimpleError("ERR wrong number of arguments for 'zrange' command")
		}

		var left, err1 = strconv.Atoi(cmd.Args[1])
		var right, err2 = strconv.Atoi(cmd.Args[2])
		if err1 != nil || err2 != nil {
			return resp.SimpleError("ERR value is not a valid integer")
		}

		var list []string = dataStore.ZRANGE(cmd.Args[0], left, right)
		var result resp.RESPType = resp.ArrayToBulkRESP(list)
		return result

	case "ZCARD":
		if len(cmd.Args) != 1 {
			return resp.SimpleError("ERR wrong number of arguments for 'zcard' command")
		}

		var count int = dataStore.ZCARD(cmd.Args[0])
		return resp.Integer(count)

	case "ZSCORE":
		if len(cmd.Args) != 2 {
			return resp.SimpleError("ERR wrong number of arguments for 'zscore' command")
		}

		score, err := dataStore.ZSCORE(cmd.Args[0], cmd.Args[1])
		if err != nil {
			return resp.Null{}
		}

		return resp.BulkString(score)

	default:
		return resp.SimpleError("ERR Invalid or Unsupported Command")

	}
}