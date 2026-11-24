package resp

// Helpers to convert Generic types to RESP types.


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