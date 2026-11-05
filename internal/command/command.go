package command

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

var SimpleOK []byte = resp.SimpleString.EncodeRESP("OK")
var ERRInvalidCommand []byte = resp.ErrorString.EncodeRESP("ERR Invalid Command")
var ERRWrongType []byte = resp.ErrorString.EncodeRESP("ERR Wrong Type")

func ExecuteCommand(conn net.Conn, cmd *resp.Command, dataStore *store.Store) []byte {
	switch cmd.Com {

	case "PING":
		{
			return resp.SimpleString.EncodeRESP("PONG")
		}

	case "ECHO":
		{
			var result []byte
			if len(cmd.Args) != 1 {
				result = ERRInvalidCommand
			} else {
				result = resp.BulkString.EncodeRESP(cmd.Args[0])
			}
			return result
		}

	case "SET":
		{
			switch len(cmd.Args) {

			case 2:
				dataStore.Set(cmd.Args[0], cmd.Args[1])
				return SimpleOK

			case 4:
				{
					var expiryType string = strings.ToUpper(cmd.Args[2])
					if expiryType != "EX" && expiryType != "PX" {
						return ERRInvalidCommand
					}

					expiryTimeMS, err := strconv.Atoi(cmd.Args[3])
					if err != nil {
						return ERRInvalidCommand
					}
					if expiryType == "EX" {
						expiryTimeMS *= 1000
					}

					var expiryDuration = time.Duration(expiryTimeMS) * time.Millisecond

					var key string = cmd.Args[0]

					dataStore.Set(key, cmd.Args[1])

					go func() {
						time.Sleep(expiryDuration)
						dataStore.Delete(key)
					}()

					return SimpleOK
				}

			default:
				return ERRInvalidCommand

			}
		}

	case "GET":
		{
			if len(cmd.Args) != 1 {
				return ERRInvalidCommand
			}

			value, ok := dataStore.Get(cmd.Args[0])

			var response []byte
			if !ok {
				response = resp.BulkString.EncodeRESP("<nil>")
			} else {
				response = resp.BulkString.EncodeRESP(value)
			}
			return response
		}

	case "RPUSH":
		{
			if len(cmd.Args) < 2 {
				return ERRInvalidCommand
			}

			len := dataStore.RPush(cmd.Args[0], cmd.Args[1:])
			if len == -1 {
				return ERRWrongType
			}

			var response []byte = resp.Integer.EncodeRESP(len)
			return response
		}

	case "LPUSH":
		{
			if len(cmd.Args) < 2 {
				return ERRInvalidCommand
			}

			len := dataStore.LPush(cmd.Args[0], cmd.Args[1:])
			if len == -1 {
				return ERRWrongType
			}

			var response []byte = resp.Integer.EncodeRESP(len)
			return response
		}

	case "LRANGE":
		{
			if len(cmd.Args) != 3 {
				return ERRInvalidCommand
			}

			left, err1 := strconv.Atoi(cmd.Args[1])
			right, err2 := strconv.Atoi(cmd.Args[2])
			if err1 != nil || err2 != nil {
				return ERRInvalidCommand
			}

			var list []string = dataStore.LRange(cmd.Args[0], left, right)

			response := resp.Array.EncodeRESP(list)
			return response
		}

	case "LLEN":
		{
			if len(cmd.Args) != 1 {
				return ERRInvalidCommand
			}

			var len int = dataStore.LLen(cmd.Args[0])
			response := resp.Integer.EncodeRESP(len)
			return response
		}

	case "LPOP":
		{
			switch len(cmd.Args) {
			case 1:
				{
					value, ok := dataStore.LPop(cmd.Args[0])

					var response []byte
					if !ok {
						response = resp.BulkString.EncodeRESP("<nil>")
					} else {
						response = resp.BulkString.EncodeRESP(value)
					}
					return response
				}
			case 2:
				{
					var popCount, err = strconv.Atoi(cmd.Args[1])
					if err != nil || popCount < 0 {
						return ERRInvalidCommand
					}
					var response []byte

					values, ok := dataStore.LPopN(cmd.Args[0], popCount)
					if !ok {
						response = resp.Array.EncodeRESP(nil)
					} else {
						response = resp.Array.EncodeRESP(values)
					}
					return response
				}
			default:
				{
					return ERRInvalidCommand
				}
			}
		}

	case "BLPOP":
		{
			if len(cmd.Args) != 2 {
				return ERRInvalidCommand
			}

			timeout, err := strconv.ParseFloat(cmd.Args[1], 64)
			if err != nil || timeout < 0 {
				return ERRInvalidCommand
			}
			var key string = cmd.Args[0]

			value, ok := dataStore.BLPop(key, timeout)
			if !ok {
				var lst []string = nil
				response := resp.Array.EncodeRESP(lst)
				return response
			}

			response := resp.Array.EncodeRESP([]string{key, value})
			return response
		}

	case "TYPE":
		{
			if len(cmd.Args) != 1 {
				return ERRInvalidCommand
			}

			var key string = cmd.Args[0]
			var typeStr string = dataStore.TypeOf(key)

			response := resp.SimpleString.EncodeRESP(typeStr)
			return response
		}

	case "XADD":
		{
			if len(cmd.Args) < 3 || len(cmd.Args)%2 != 0 {
				return ERRInvalidCommand
			}

			var key string = cmd.Args[0]
			var ID string = cmd.Args[1]
			var fields map[string]string = make(map[string]string)
			for i := 2; i < len(cmd.Args); i += 2 {
				fields[cmd.Args[i]] = cmd.Args[i+1]
			}

			newID, err := dataStore.AddStreamEntry(key, ID, fields)
			if err != nil {
				return resp.ErrorString.EncodeRESP(err.Error())
			}

			response := resp.BulkString.EncodeRESP(newID)
			return response
		}

	case "XRANGE":
		{
			if len(cmd.Args) != 3 {
				return resp.ErrorString.EncodeRESP("ERR Invalid Arguments for XRANGE command")
			}

			var key string = cmd.Args[0]
			var start string = cmd.Args[1]
			var end string = cmd.Args[2]

			startMST, startSeq, err1 := validateStreamID(start)
			endMST, endSeq, err2 := validateStreamID(end)
			if err1 != nil {
				return resp.ErrorString.EncodeRESP(err1.Error())
			}
			if err2 != nil {
				return resp.ErrorString.EncodeRESP(err2.Error())
			}
			if startSeq == -1 {
				startSeq = 0
			}

			entries, err := dataStore.XRange(key, startMST, startSeq, endMST, endSeq)
			if err != nil {
				return resp.ErrorString.EncodeRESP(err.Error())
			}

			response := resp.EncodeRESPStream(entries)
			return response
		}
	default:
		{
			return ERRInvalidCommand
		}

	}
}

func validateStreamID(ID string) (int64, int64, error) {
	var mst, seq int64
	if strings.Contains(ID, "-") {
		var parts []string = strings.Split(ID, "-")
		if len(parts) != 2 {
			return -1, -1, errors.New("ERR Invalid Arguments for XRANGE command")
		}
		val1, err1 := strconv.ParseInt(parts[0], 10, 64)
		val2, err2 := strconv.ParseInt(parts[1], 10, 64)
		if err1 != nil || err2 != nil {
			return -1, -1, errors.New("ERR Invalid Arguments for XRANGE command")
		}
		mst = val1
		seq = val2
	} else {
		val, err := strconv.ParseInt(ID, 10, 64)
		if err != nil {
			return -1, -1, errors.New("ERR Invalid Arguments for XRANGE command")
		}
		mst = val
		seq = -1
	}
	return mst, seq, nil
}
