package handlers

import (
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"

	Error "github.com/codecrafters-io/redis-starter-go/internal/errors"
)

func String(cmd *resp.Command, dataStore *store.Store) resp.RESPType {
	switch cmd.Com {

	case "SET":
		switch len(cmd.Args) {

		case 2:
			err := dataStore.Set(cmd.Args[0], cmd.Args[1])
			if err != nil {
				return resp.SimpleError(err.Error())
			}
			return resp.SimpleString("OK")

		case 4:
			var expiryType string = strings.ToUpper(cmd.Args[2])
			expiryTimeMS, err := strconv.Atoi(cmd.Args[3])
			if err != nil {
				return resp.SimpleError("ERR Invalid or Unsupported Arguments for SET command")
			}
			if expiryType == "EX" {
				expiryTimeMS *= 1000
			}

			var expiryDuration = time.Duration(expiryTimeMS) * time.Millisecond

			var key string = cmd.Args[0]

			err = dataStore.Set(key, cmd.Args[1])
			if err != nil {
				return resp.SimpleError(err.Error())
			}

			go func() {
				time.Sleep(expiryDuration)
				dataStore.Delete(key)
			}()

			return resp.SimpleString("OK")
		}

	case "GET":
		value, err := dataStore.Get(cmd.Args[0])
		if err != nil {
			return resp.SimpleError(err.Error())
		}

		var result resp.RESPType
		if value == "" {
			result = resp.Null{}
		} else {
			result = resp.BulkString(value)
		}
		return result

	case "INCR":
		newValue, err := dataStore.Incr(cmd.Args[0])
		if err != nil {
			return resp.SimpleError(err.Error())
		}
		return resp.Integer(newValue)
	}

	return resp.SimpleError(Error.ERR_INVALID.Error())
}
