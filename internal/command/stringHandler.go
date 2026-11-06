package command

import (
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func stringHandler(cmd *resp.Command, dataStore *store.Store) resp.RESPType {
	switch cmd.Com {

	case "SET":
		{
			switch len(cmd.Args) {

			case 2:
				err := dataStore.Set(cmd.Args[0], cmd.Args[1])
				if err != nil {
					return resp.ErrorString(err.Error())
				}
				return resp.SimpleString("OK")

			case 4:
				{
					var expiryType string = strings.ToUpper(cmd.Args[2])
					if expiryType != "EX" && expiryType != "PX" {
						return resp.ErrorString("ERR Invalid or Unsupported Arguments for SET command")
					}

					expiryTimeMS, err := strconv.Atoi(cmd.Args[3])
					if err != nil {
						return resp.ErrorString("ERR Invalid or Unsupported Arguments for SET command")
					}
					if expiryType == "EX" {
						expiryTimeMS *= 1000
					}

					var expiryDuration = time.Duration(expiryTimeMS) * time.Millisecond

					var key string = cmd.Args[0]

					err = dataStore.Set(key, cmd.Args[1])
					if err != nil {
						return resp.ErrorString(err.Error())
					}

					go func() {
						time.Sleep(expiryDuration)
						dataStore.Delete(key)
					}()

					return resp.SimpleString("OK")
				}

			default:
				return resp.ErrorString("ERR Invalid or Unsupported Arguments for SET command")

			}
		}

	case "GET":
		{
			if len(cmd.Args) != 1 {
				return resp.ErrorString("ERR Invalid command syntax for GET command")
			}

			value, err := dataStore.Get(cmd.Args[0])
			if err != nil {
				return resp.ErrorString(err.Error())
			}

			var result resp.RESPType
			if value == "" {
				result = resp.Null{}
			} else {
				result = resp.BulkString(value)
			}
			return result
		}

	default:
		return resp.ErrorString("ERR Invalid or Unsupported Command")

	}
}
