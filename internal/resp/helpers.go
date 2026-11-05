package resp

func BuildRESPArray(values []string) Array {
	var respArray Array
	for _, value := range values {
		respArray = append(respArray, BulkString(value))
	}
	return respArray
}