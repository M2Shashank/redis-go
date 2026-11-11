package resp

// Helpers to convert Store types to RESP types.

import (
    "github.com/codecrafters-io/redis-starter-go/internal/store"
)

// ArrayToBulkRESP converts a slice of strings to a RESP array of bulk strings.
func ArrayToBulkRESP(values []string) RESPType {
	if values == nil {
		return Array(nil)
	}

	var respArray Array = Array{}
	for _, value := range values {
		respArray = append(respArray, BulkString(value))
	}
	return respArray
}

func CommandToRESP(cmd *Command) RESPType {
	if cmd == nil {
		return Array(nil)
	}
	var respArray Array = Array{}
	respArray = append(respArray, BulkString(cmd.Com))
	for _, arg := range cmd.Args {
		respArray = append(respArray, BulkString(arg))
	}
	return respArray
}

func StreamEntriesToRESP(entries []*store.StreamEntry) RESPType {
	if len(entries) == 0 {
		return Array(nil)
	}
    result := make(Array, 0, len(entries))

    for _, entry := range entries {
        fields := make(Array, 0, len(entry.Fields)*2)

        for field, value := range entry.Fields {
            fields = append(fields,
                BulkString(field),
                BulkString(value),
            )
        }

        streamEntry := Array{
            BulkString(entry.ID),
            fields,
        }

        result = append(result, streamEntry)
    }

    return result
}

func ModStreamToRESP(input *store.ModStream) RESPType {
	if input == nil || len(input.Entries) == 0 {
		return Array(nil)
	}

	result := Array {
		BulkString(input.Key),
		StreamEntriesToRESP(input.Entries),
	}

	return result
}

func ModStreamsToRESP(inputs []*store.ModStream) RESPType {
	if len(inputs) == 0 {
		return Array(nil)
	}

	result := make(Array, 0, len(inputs))
	for _, input := range inputs {
		result = append(result, ModStreamToRESP(input))
	}

	return result
}
