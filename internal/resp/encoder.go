package resp

import (
	"fmt"
	"strings"
)

type EncoderType int 

const (
	SimpleString EncoderType = iota
 	ErrorString
	Integer
	BulkString
	Array
)

func (e EncoderType) EncodeRESP(value any) []byte {
	switch e {
		case SimpleString:
			return []byte("+" + value.(string) + "\r\n")
		case ErrorString:
			return []byte("-" + value.(string) + "\r\n")
		case Integer:
			return fmt.Appendf([]byte{}, ":%d\r\n", value.(int))
		case BulkString:
			var str string = value.(string)
			if str == "<nil>" {
				return []byte("$-1\r\n")
			}
			var resp string = fmt.Sprintf("$%d\r\n%s\r\n", len(str), str)
			return []byte(resp)
		case Array:
			var list []string = value.([]string)
			if list == nil {
				return []byte("*-1\r\n")
			}
			var resp strings.Builder
			fmt.Fprintf(&resp, "*%d\r\n", len(list))
			for _, v := range list {
				fmt.Fprintf(&resp, "$%d\r\n%s\r\n", len(v), v)
			}
			return []byte(resp.String())
		default:
			return []byte("-ERR Unknown Encoder Type\r\n")
	}
}