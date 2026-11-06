package command

import (
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func streamHandler(cmd *resp.Command, dataStore *store.Store) resp.RESPType {
	switch cmd.Com {

	case "XADD":
		{
			if len(cmd.Args) < 3 || len(cmd.Args)%2 != 0 {
				return resp.SimpleError("ERR Invalid Arguments for XADD command")
			}

			var key string = cmd.Args[0]
			var ID string = cmd.Args[1]
			var fields map[string]string = make(map[string]string)
			for i := 2; i < len(cmd.Args); i += 2 {
				fields[cmd.Args[i]] = cmd.Args[i+1]
			}

			newID, err := dataStore.AddStreamEntry(key, ID, fields)
			if err != nil {
				return resp.SimpleError(err.Error())
			}

			return resp.BulkString(newID)
		}

	case "XRANGE":
		{
			if len(cmd.Args) != 3 {
				return resp.SimpleError("ERR Invalid Arguments for XRANGE command")
			}

			var key string = cmd.Args[0]
			var start string = cmd.Args[1]
			var end string = cmd.Args[2]

			startMST, startSeq, err1 := ValidateStreamID(start, "XRANGE")
			endMST, endSeq, err2 := ValidateStreamID(end, "XRANGE")
			if err1 != nil {
				return resp.SimpleError(err1.Error())
			}
			if err2 != nil {
				return resp.SimpleError(err2.Error())
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
				return resp.SimpleError(err.Error())
			}

			response := StreamEntriesToRESP(entries)
			return response
		}

	case "XREAD":
		{
			var length int = len(cmd.Args)
			if length < 3 || length%2 != 1 {
				return resp.SimpleError("ERR Invalid Arguments for XREAD command")
			}

			var expiryTime int64 = -1
			if strings.ToUpper(cmd.Args[0]) == "BLOCK" {
				time, err := strconv.ParseInt(cmd.Args[1], 10, 64)
				if err != nil || time < 0 {
					return resp.SimpleError("ERR Invalid Arguments for XREAD command")
				}
				expiryTime = time
				cmd.Args = cmd.Args[2:]
				length -= 2
			}

			if length < 3 || length%2 != 1 {
				return resp.SimpleError("ERR Invalid Arguments for XREAD command")
			}

			var str string = strings.ToUpper(cmd.Args[0])
			if str != "STREAMS" {
				return resp.SimpleError("ERR Invalid Arguments for XREAD command")
			}

			var keys []string = cmd.Args[1 : (length/2)+1]
			var ids []string = cmd.Args[(length/2)+1:]

			var mst, seq []int64 = make([]int64, len(ids)), make([]int64, len(ids))

			// validate all ids
			for index, id := range ids {
				if id == "$" {
					mst[index], seq[index] = -1, -1
					continue
				}
				mstNum, seqNum, err := ValidateStreamID(id, "XREAD")
				if err != nil {
					return resp.SimpleError(err.Error())
				}
				mst[index], seq[index] = mstNum, seqNum
			}

			result, empty := dataStore.XReadStreams(keys, mst, seq)
			if empty && expiryTime != -1 {
				resultNew, err := dataStore.BXReadStreams(keys, mst, seq, expiryTime)
				if err != nil {
					return resp.SimpleError(err.Error())
				}
				result = resultNew
			}

			response := ModStreamsToRESP(result)
			return response
		}

	default:
		return resp.SimpleError("ERR Invalid or Unsupported Command")

	}
}
