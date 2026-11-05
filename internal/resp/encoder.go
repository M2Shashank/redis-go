package resp

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/store"
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

func EncodeRESPStream(entry []*store.StreamEntry) []byte {
	var resp strings.Builder
	fmt.Fprintf(&resp, "*%d\r\n", len(entry))
	for _, e := range entry {
		fmt.Fprintf(&resp, "*2\r\n")
		fmt.Fprintf(&resp, "$%d\r\n%s\r\n", len(e.ID), e.ID)
		fmt.Fprintf(&resp, "*%d\r\n", 2*len(e.Fields))
		for field, value := range e.Fields {
			fmt.Fprintf(&resp, "$%d\r\n%s\r\n", len(field), field)
			fmt.Fprintf(&resp, "$%d\r\n%s\r\n", len(value), value)
		}
	}
	return []byte(resp.String())
}