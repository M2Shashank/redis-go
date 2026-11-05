package resp

import (
	"fmt"
	"strings"
)

func Encode(value RESPType) []byte {
	var builder strings.Builder
	EncodeRESP(&builder, value)
	return []byte(builder.String())
}

func EncodeRESP(builder *strings.Builder, value RESPType) {
	switch v := value.(type) {

    case SimpleString:
        fmt.Fprintf(builder, "+%s\r\n", string(v))

    case ErrorString:
        fmt.Fprintf(builder, "-%s\r\n", string(v))

    case Integer:
        fmt.Fprintf(builder, ":%d\r\n", int64(v))

    case BulkString:
        str := string(v)
        fmt.Fprintf(builder, "$%d\r\n%s\r\n", len(str), str)

    case Null:
        builder.WriteString("$-1\r\n")

    case Array:
		if v == nil {
			fmt.Fprintf(builder, "*-1\r\n")
			return
		}
		
        fmt.Fprintf(builder, "*%d\r\n", len(v))

        for _, item := range v {
            EncodeRESP(builder, item)
        }

    default:
        fmt.Fprintf(builder, "-ERR Unsupported RESP type\r\n")
    }
}