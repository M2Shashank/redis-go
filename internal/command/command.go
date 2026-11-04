package command

import (
	"fmt"
	"net"
	"time"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"github.com/codecrafters-io/redis-starter-go/internal/protocol"
)

func ExecuteCommand(conn net.Conn, cmd *protocol.Command, store *store.Store) {
	switch cmd.Com {

	case "PING":
		{
			conn.Write([]byte("+PONG\r\n"))
		}

	case "ECHO":
		{
			if len(cmd.Args) != 1 {
				conn.Write([]byte("-ERR Invalid Command\r\n"))
			} else {
				var arg string = cmd.Args[0]
				var resp string = fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
				conn.Write([]byte(resp))
			}
		}

	case "SET":
		{
			switch len(cmd.Args) {

			case 2:
				{
					store.Set(cmd.Args[0], cmd.Args[1])
					conn.Write([]byte("+OK\r\n"))
				}

			case 4:
				{
					var expiryType string = strings.ToUpper(cmd.Args[2])
					if expiryType != "EX" && expiryType != "PX" {
						conn.Write([]byte("-ERR Invalid Command\r\n"))
						return
					}

					expiryTimeMS, err := strconv.Atoi(cmd.Args[3])
					if err != nil {
						conn.Write([]byte("-ERR Invalid Command\r\n"))
						return
					}
					if expiryType == "EX" {
						expiryTimeMS *= 1000
					}

					var expiryDuration = time.Duration(expiryTimeMS) * time.Millisecond

					var key string = cmd.Args[0]

					store.Set(key, cmd.Args[1])

					go func() {
						time.Sleep(expiryDuration)
						store.Delete(key)
					}()

					conn.Write([]byte("+OK\r\n"))
				}

			default:
				conn.Write([]byte("-ERR Invalid Command\r\n"))

			}
		}

	case "GET":
		{
			if len(cmd.Args) != 1 {
				conn.Write([]byte("-ERR Invalid Command\r\n"))
				return
			}

			value, ok := store.Get(cmd.Args[0])

			if !ok {
				conn.Write([]byte("$-1\r\n"))
			} else {
				var resp string = fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
				conn.Write([]byte(resp))
			}
		}

	default:
		{
			conn.Write([]byte("-ERR Unknown Command\r\n"))
		}

	}
}