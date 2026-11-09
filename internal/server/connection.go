package server

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func HandleConn(client *Client, s *store.Store) {
	defer func() {
		client.Server.removeClient(client)
		(client.Conn).Close()
	}()
	var buff []byte = make([]byte, 1024)
	var readBuffer []byte
	var server *Server = client.Server

	var masterId string = server.replicationInfo["master_replid"]
	var masterOffset string = server.replicationInfo["master_repl_offset"]

	for {
		n, err := (client.Conn).Read(buff)
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			break
		}
		fmt.Printf("Read %d bytes from connection: %s\n", n, string(buff[:n]))

		readBuffer = append(readBuffer, buff[:n]...)

		for {
			n, cmd := resp.ParseRESPCommand(readBuffer)
			if cmd == nil || n == 0 {
				break
			}
			readBuffer = readBuffer[n:]

			var result resp.RESPType
			switch cmd.Com {
			case "WATCH", "UNWATCH", "MULTI", "EXEC", "DISCARD":
				result = client.TransactionHandler(cmd, s)
			case "INFO":
				result = InfoHandler(cmd, client.Server)
			case "REPLCONF":
				result = resp.SimpleString("OK")
			case "PSYNC":
				str := fmt.Sprintf("FULLRESYNC %s %s", masterId, masterOffset)
				result = resp.SimpleString(str)
			default:
				if client.inTransaction {
					client.queuedCmds.Push(cmd)
					result = resp.SimpleString("QUEUED")
				} else {
					result = command.ExecuteCommand(cmd, s)
				}
			}

			response := resp.Encode(result)

			(client.Conn).Write(response)
		}
	}
}
