package handlers

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"

	Error "github.com/codecrafters-io/redis-starter-go/internal/errors"
)

func Zset(cmd *resp.Command, dataStore *store.Store) resp.RESPType {
	switch cmd.Com {
	case "ZADD":
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
				ans++
			}
		}
		return resp.Integer(ans)

	case "ZRANK":
		index, err := dataStore.ZRANK(cmd.Args[0], cmd.Args[1])
		if err != nil {
			return resp.Null{}
		}

		return resp.Integer(index)

	case "ZRANGE":
		var left, err1 = strconv.Atoi(cmd.Args[1])
		var right, err2 = strconv.Atoi(cmd.Args[2])
		if err1 != nil || err2 != nil {
			return resp.SimpleError("ERR value is not a valid integer")
		}

		var list []string = dataStore.ZRANGE(cmd.Args[0], left, right)
		var result resp.RESPType = resp.ArrayToBulkRESP(list)
		return result

	case "ZCARD":
		var count int = dataStore.ZCARD(cmd.Args[0])
		return resp.Integer(count)

	case "ZSCORE":
		score, err := dataStore.ZSCORE(cmd.Args[0], cmd.Args[1])
		if err != nil {
			return resp.Null{}
		}

		return resp.BulkString(score)

	case "ZREM":
		var ans int = 0
		for i := 1; i < len(cmd.Args); i++ {
			removed, err := dataStore.ZREM(cmd.Args[0], cmd.Args[i])
			if err != nil {
				return resp.SimpleError(err.Error())
			}

			if removed {
				ans++
			}
		}
		return resp.Integer(ans)
	}

	return resp.SimpleError(Error.ERR_INVALID.Error())
}
