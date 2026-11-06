package command

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func listHandler(cmd *resp.Command, dataStore *store.Store) resp.RESPType {
	switch cmd.Com {

	case "RPUSH":
		{
			if len(cmd.Args) < 2 {
				return resp.ErrorString("ERR Invalid command syntax for RPUSH command")
			}

			len, err := dataStore.RPush(cmd.Args[0], cmd.Args[1:])
			if err != nil {
				return resp.ErrorString(err.Error())
			}

			var result resp.RESPType = resp.Integer(len)
			return result
		}

	case "LPUSH":
		{
			if len(cmd.Args) < 2 {
				return resp.ErrorString("ERR Invalid command syntax for LPUSH command")
			}

			len, err := dataStore.LPush(cmd.Args[0], cmd.Args[1:])
			if err != nil {
				return resp.ErrorString(err.Error())
			}

			var result resp.RESPType = resp.Integer(len)
			return result
		}

	case "LRANGE":
		{
			if len(cmd.Args) != 3 {
				return resp.ErrorString("ERR Invalid command syntax for LRANGE command")
			}

			left, err1 := strconv.Atoi(cmd.Args[1])
			right, err2 := strconv.Atoi(cmd.Args[2])
			if err1 != nil || err2 != nil {
				return resp.ErrorString("ERR Invalid or Unsupported Arguments for LRANGE command")
			}

			var list []string = dataStore.LRange(cmd.Args[0], left, right)
			var respArray resp.RESPType = BulkArrayToRESP(list)

			return respArray
		}

	case "LLEN":
		{
			if len(cmd.Args) != 1 {
				return resp.ErrorString("ERR Invalid command syntax for LLEN command")
			}

			len, err := dataStore.LLen(cmd.Args[0])
			if err != nil {
				return resp.ErrorString(err.Error())
			}

			return resp.Integer(len)
		}

	case "LPOP":
		{
			switch len(cmd.Args) {
			case 1:
				{
					value, err := dataStore.LPop(cmd.Args[0])

					var response resp.RESPType
					if err != nil {
						response = resp.Null{}
					} else {
						response = resp.BulkString(value)
					}
					return response
				}
			case 2:
				{
					var popCount, err = strconv.Atoi(cmd.Args[1])
					if err != nil || popCount < 0 {
						return resp.ErrorString("ERR Invalid or Unsupported Arguments for LPOP command")
					}

					values, err := dataStore.LPopN(cmd.Args[0], popCount)
					if err != nil {
						return resp.ErrorString(err.Error())
					}

					response := BulkArrayToRESP(values)
					return response
				}
			default:
				{
					return resp.ErrorString("ERR Invalid command syntax for LPOP command")
				}
			}
		}

	case "BLPOP":
		{
			if len(cmd.Args) != 2 {
				return resp.ErrorString("ERR Invalid command syntax for BLPOP command")
			}

			timeout, err := strconv.ParseFloat(cmd.Args[1], 64)
			if err != nil || timeout < 0 {
				return resp.ErrorString("ERR Invalid command syntax for BLPOP command")
			}
			var key string = cmd.Args[0]

			value, err := dataStore.BLPop(key, timeout)
			if err != nil {
				return resp.ErrorString(err.Error())
			}

			var arr []string = []string{key, value}
			var respArray resp.RESPType = BulkArrayToRESP(arr)

			return respArray
		}

	default:
		return resp.ErrorString("ERR Invalid or Unsupported Command")

	}
}
