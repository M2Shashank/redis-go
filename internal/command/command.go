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
var ERRInvalidCommand resp.RESPType = resp.ErrorString("ERR Invalid Command")
var ERRWrongType resp.RESPType = resp.ErrorString("ERR Wrong Type")

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
				result = ERRInvalidCommand
			} else {
				result = resp.BulkString(cmd.Args[0])
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

			var result resp.RESPType
			if !ok {
				result = resp.Null{}
			} else {
				result = resp.BulkString(value)
			}
			return result
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

			var result resp.RESPType = resp.Integer(len)
			return result
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


			var result resp.RESPType = resp.Integer(len)
			return result
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
			var respArray resp.Array = resp.BuildRESPArray(list)

			return respArray
		}

	case "LLEN":
		{
			if len(cmd.Args) != 1 {
				return ERRInvalidCommand
			}

			var len int = dataStore.LLen(cmd.Args[0])
			return resp.Integer(len)
		}

	case "LPOP":
		{
			switch len(cmd.Args) {
			case 1:
				{
					value, ok := dataStore.LPop(cmd.Args[0])

					var response resp.RESPType
					if !ok {
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
						return ERRInvalidCommand
					}
					var response resp.RESPType

					values, ok := dataStore.LPopN(cmd.Args[0], popCount)
					if !ok {
						response = resp.Array(nil)
					} else {
						response = resp.BuildRESPArray(values)
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
				return resp.Array(nil)
			}

			var arr []string = []string{key, value}
			var respArray resp.Array = resp.BuildRESPArray(arr)

			return respArray
		}

	case "TYPE":
		{
			if len(cmd.Args) != 1 {
				return ERRInvalidCommand
			}

			var key string = cmd.Args[0]
			var typeStr string = dataStore.TypeOf(key)

			return resp.BulkString(typeStr)
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

			startMST, startSeq, err1 := validateStreamID(start)
			endMST, endSeq, err2 := validateStreamID(end)
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
	default:
		{
			return ERRInvalidCommand
		}

	}
}
