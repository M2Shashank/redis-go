package command

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/protocol"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func ExecuteCommand(conn net.Conn, cmd *protocol.Command, dataStore *store.Store) {
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
					dataStore.Set(cmd.Args[0], cmd.Args[1])
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

					dataStore.Set(key, cmd.Args[1])

					go func() {
						time.Sleep(expiryDuration)
						dataStore.Delete(key)
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

			value, ok := dataStore.Get(cmd.Args[0])

			if !ok {
				conn.Write([]byte("$-1\r\n"))
			} else {
				var resp string = fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
				conn.Write([]byte(resp))
			}
		}

	case "RPUSH":
		{
			if len(cmd.Args) < 2 {
				conn.Write([]byte("-ERR Invalid Command\r\n"))
				return
			}

			var len int = 0
			for _, value := range cmd.Args[1:] {
				len = dataStore.RPush(cmd.Args[0], value)

				if len == -1 {
					conn.Write([]byte("-ERR Wrong Type\r\n"))
					return
				}
			}

			var resp = fmt.Sprintf(":%d\r\n", len)
			conn.Write([]byte(resp))
		}

	case "LPUSH":
		{
			if len(cmd.Args) < 2 {
				conn.Write([]byte("-ERR Invalid Command\r\n"))
				return
			}

			var len int = 0
			for _, value := range cmd.Args[1:] {
				len = dataStore.LPush(cmd.Args[0], value)

				if len == -1 {
					conn.Write([]byte("-ERR Wrong Type\r\n"))
					return
				}
			}

			var resp = fmt.Sprintf(":%d\r\n", len)
			conn.Write([]byte(resp))
		}

	case "LRANGE":
		{
			if len(cmd.Args) != 3 {
				conn.Write([]byte("-ERR Invalid Command\r\n"))
				return
			}

			left, err1 := strconv.Atoi(cmd.Args[1])
			right, err2 := strconv.Atoi(cmd.Args[2])
			if err1 != nil || err2 != nil {
				conn.Write([]byte("-ERR Invalid Command\r\n"))
				return
			}

			var list []string = dataStore.LRange(cmd.Args[0], left, right)

			if list == nil {
				conn.Write([]byte("*0\r\n"))
				return
			}

			var resp string = fmt.Sprintf("*%d\r\n", len(list))
			for _, value := range list {
				resp += fmt.Sprintf("$%d\r\n%s\r\n", len(value), value)
			}
			conn.Write([]byte(resp))
		}

	case "LLEN":
		{
			if len(cmd.Args) != 1 {
				conn.Write([]byte("-ERR Invalid Command\r\n"))
				return
			}

			var len int = dataStore.LLen(cmd.Args[0])
			var resp = fmt.Sprintf(":%d\r\n", len)
			conn.Write([]byte(resp))
		}

	default:
		{
			conn.Write([]byte("-ERR Unknown Command\r\n"))
		}

	}
}
