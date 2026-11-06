package command

import (
    "github.com/codecrafters-io/redis-starter-go/internal/resp"
    "github.com/codecrafters-io/redis-starter-go/internal/store"
)

func BulkArrayToRESP(values []string) resp.RESPType {
	if values == nil {
		return resp.Array(nil)
	}

	var respArray resp.Array = resp.Array{}
	for _, value := range values {
		respArray = append(respArray, resp.BulkString(value))
	}
	return respArray
}

func StreamEntriesToRESP(entries []*store.StreamEntry) resp.RESPType {
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
