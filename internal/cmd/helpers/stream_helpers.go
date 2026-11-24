package helpers

import (
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"

	store_model "github.com/codecrafters-io/redis-starter-go/internal/store/models"
	Error "github.com/codecrafters-io/redis-starter-go/internal/errors"
)

func ValidateStreamID(ID string, command string) (int64, int64, error) {
	var mst, seq int64
	if command == "XRANGE" {
		if ID == "-" {
			return -1, -1, nil
		}
		if ID == "+" {
			return -2, -2, nil
		}
	}

	if strings.Contains(ID, "-") {
		var parts []string = strings.Split(ID, "-")
		if len(parts) != 2 {
			return -1, -1, Error.ERR_InvalidArguments(command)
		}
		val1, err1 := strconv.ParseInt(parts[0], 10, 64)
		val2, err2 := strconv.ParseInt(parts[1], 10, 64)
		if err1 != nil || err2 != nil {
			return -1, -1, Error.ERR_InvalidArguments(command)
		}
		mst = val1
		seq = val2
	} else {
		val, err := strconv.ParseInt(ID, 10, 64)
		if err != nil {
			return -1, -1, Error.ERR_InvalidArguments(command)
		}

		mst = val
		seq = -1
	}
	return mst, seq, nil
}




func StreamEntriesToRESP(entries []*store_model.StreamEntry) resp.RESPType {
	if len(entries) == 0 {
		return resp.Array(nil)
	}
	result := make(resp.Array, 0, len(entries))

	for _, entry := range entries {
		fields := make(resp.Array, 0, len(entry.Fields)*2)

		for field, value := range entry.Fields {
			fields = append(fields,
				resp.BulkString(field),
				resp.BulkString(value),
			)
		}

		streamEntry := resp.Array{
			resp.BulkString(entry.ID),
			fields,
		}

		result = append(result, streamEntry)
	}

	return result
}

func ModStreamToRESP(input *store_model.ModStream) resp.RESPType {
	if input == nil || len(input.Entries) == 0 {
		return resp.Array(nil)
	}

	result := resp.Array{
		resp.BulkString(input.Key),
		StreamEntriesToRESP(input.Entries),
	}

	return result
}

func ModStreamsToRESP(inputs []*store_model.ModStream) resp.RESPType {
	if len(inputs) == 0 {
		return resp.Array(nil)
	}

	result := make(resp.Array, 0, len(inputs))
	for _, input := range inputs {
		result = append(result, ModStreamToRESP(input))
	}

	return result
}