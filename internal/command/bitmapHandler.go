package command

import (
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func bitmapHandler(cmd *resp.Command, dataStore *store.Store) resp.RESPType {
	switch cmd.Com {
	case "SETBIT":
		if len(cmd.Args) != 3 {
			return resp.SimpleError("ERR wrong number of arguments for 'setbit' command")
		}

		key, offsetStr, valueStr := cmd.Args[0], cmd.Args[1], cmd.Args[2]
		offset, err1 := strconv.Atoi(offsetStr)
		value, err2 := strconv.Atoi(valueStr)
		if err1 != nil || err2 != nil || (value != 0 && value != 1) {
			return resp.SimpleError("ERR value is not an integer or out of range")
		}

		prevValue, err := dataStore.SetBit(key, offset, value)
		if err != nil {
			return resp.SimpleError(err.Error())
		}
		return resp.Integer(prevValue)

	case "GETBIT":
		if len(cmd.Args) != 2 {
			return resp.SimpleError("ERR wrong number of arguments for 'getbit' command")
		}

		key, offsetStr := cmd.Args[0], cmd.Args[1]
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return resp.SimpleError("ERR value is not an integer or out of range")
		}

		value := dataStore.GetBit(key, offset)
		return resp.Integer(value)

	case "STRLEN":
		if len(cmd.Args) != 1 {
			return resp.SimpleError("ERR wrong number of arguments for 'strlen' command")
		}

		value := dataStore.StrLen(cmd.Args[0])
		return resp.Integer(value)

	case "BITCOUNT":
		if len(cmd.Args) < 1 {
			return resp.SimpleError("ERR wrong number of arguments for 'bitcount' command")
		}

		var start, end int = 0, -1 // Default values for start and end
		var err error
		if len(cmd.Args) >= 2 {
			start, err = strconv.Atoi(cmd.Args[1])
			if err != nil {
				return resp.SimpleError("ERR value is not an integer or out of range")
			}
		}
		if len(cmd.Args) == 3 {
			end, err = strconv.Atoi(cmd.Args[2])
			if err != nil {
				return resp.SimpleError("ERR value is not an integer or out of range")
			}
		}

		value := dataStore.BitCount(cmd.Args[0], start, end)
		return resp.Integer(value)

	default:
		return resp.SimpleError("ERR Invalid or Unsupported Command")
	}
}