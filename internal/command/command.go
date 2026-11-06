package command

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

var SimpleOK resp.RESPType = resp.SimpleString("OK")

func ExecuteCommand(conn net.Conn, cmd *resp.Command, dataStore *store.Store) resp.RESPType {
	switch cmd.Com {

	case "PING":
		{
			return resp.SimpleString("PONG")
		}

	case "ECHO":
		{
			var result resp.RESPType
			if len(cmd.Args) != 1 {
				result = resp.ErrorString("ERR Invalid command syntax for ECHO command")
			} else {
				result = resp.BulkString(cmd.Args[0])
			}
			return result
		}

	case "SET":
		{
			switch len(cmd.Args) {

			case 2:
				err := dataStore.Set(cmd.Args[0], cmd.Args[1])
				if err != nil {
					return resp.ErrorString(err.Error())
				}
				return SimpleOK

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

					return SimpleOK
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

	case "TYPE":
		{
			if len(cmd.Args) != 1 {
				return resp.ErrorString("ERR Invalid command syntax for TYPE command")
			}

			var key string = cmd.Args[0]
			var typeStr string = dataStore.TypeOf(key)

			return resp.SimpleString(typeStr)
		}

	case "XADD":
		{
			if len(cmd.Args) < 3 || len(cmd.Args)%2 != 0 {
				return resp.ErrorString("ERR Invalid Arguments for XADD command")
			}

			var key string = cmd.Args[0]
			var ID string = cmd.Args[1]
			var fields map[string]string = make(map[string]string)
			for i := 2; i < len(cmd.Args); i += 2 {
				fields[cmd.Args[i]] = cmd.Args[i+1]
			}

			newID, err := dataStore.AddStreamEntry(key, ID, fields)
			if err != nil {
				return resp.ErrorString(err.Error())
			}

			return resp.BulkString(newID)
		}

	case "XRANGE":
		{
			if len(cmd.Args) != 3 {
				return resp.ErrorString("ERR Invalid Arguments for XRANGE command")
			}

			var key string = cmd.Args[0]
			var start string = cmd.Args[1]
			var end string = cmd.Args[2]

			startMST, startSeq, err1 := validateStreamID(start, "XRANGE")
			endMST, endSeq, err2 := validateStreamID(end, "XRANGE")
			if err1 != nil {
				return resp.ErrorString(err1.Error())
			}
			if err2 != nil {
				return resp.ErrorString(err2.Error())
			}
			startMST = max(startMST, 0)
			startSeq = max(startSeq, 0)
			if endMST == -1 {
				endMST = 0
				endSeq = 0
			}
			if endMST == -2 {
				endMST = -1
				endSeq = -1
			}

			entries, err := dataStore.XRange(key, startMST, startSeq, endMST, endSeq)
			if err != nil {
				return resp.ErrorString(err.Error())
			}

			response := StreamEntriesToRESP(entries)
			return response
		}

	case "XREAD":
		{
			if len(cmd.Args) < 3 || len(cmd.Args)%2 != 1 {
				return resp.ErrorString("ERR Invalid Arguments for XREAD command")
			}
			if strings.ToUpper(cmd.Args[0]) != "STREAMS" {
				return resp.ErrorString("ERR Invalid Arguments for XREAD command")
			}

			var keys []string = cmd.Args[1:(len(cmd.Args)/2)+1]
			var ids []string = cmd.Args[(len(cmd.Args)/2)+1:]

			response := make(resp.Array, 0, len(keys))
			for i := 0; i < len(keys); i++ {
				key, id := keys[i], ids[i]
				startMST, startSeq, err := validateStreamID(id, "XREAD")
				if err != nil {
					return resp.ErrorString(err.Error())
				}

				entries, err := dataStore.XRange(key, startMST, startSeq, -1, -1)
				if err != nil {
					return resp.ErrorString(err.Error())
				}

				if len(entries) != 0 && entries[0].ID == id {
					entries = entries[1:]
				}

				var streamResponse resp.Array = resp.Array{
					resp.BulkString(key),
					StreamEntriesToRESP(entries),
				}
				response = append(response, streamResponse)
			}
			
			return response
		}
	default:
		{
			return resp.ErrorString("ERR Invalid or Unsupported Command")
		}

	}
}
