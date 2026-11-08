package server

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func HandleConn(conn net.Conn, s *store.Store) {
	defer conn.Close()
	var buff []byte = make([]byte, 1024)
	var readBuffer []byte

	var queuedCommands *store.Queue[*resp.Command] = store.NewQueue[*resp.Command]()
	var inQueued bool = false

	for {
		n, err := conn.Read(buff)
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			break
		}

		readBuffer = append(readBuffer, buff[:n]...)

		for {
			fmt.Println("Read buffer: ", string(readBuffer))

			n, cmd := resp.ParseRESPCommand(readBuffer)
			if cmd == nil || n == 0 {
				break
			}
			readBuffer = readBuffer[n:]

			if cmd.Com == "EXEC" {
				var results resp.Array = make([]resp.RESPType, 0, queuedCommands.Size())
				for !queuedCommands.Empty() {
					c := queuedCommands.Pop()
					result := command.ExecuteCommand(c, s)
					results = append(results, result)
				}
				response := resp.Encode(results)
				conn.Write(response)
				inQueued = false
				continue
			}

			if cmd.Com != "EXEC" && inQueued {
				queuedCommands.Push(cmd)
				response := resp.Encode(resp.SimpleString("QUEUED"))
				conn.Write(response)
				continue
			}

			result := command.ExecuteCommand(cmd, s)
			response := resp.Encode(result)

			conn.Write(response)

			if cmd.Com == "MULTI" {
				inQueued = true
			}
		}
	}
}
