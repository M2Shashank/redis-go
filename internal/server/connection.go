package server

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func HandleConn(client *Client, s *store.Store) {
	defer (client.Conn).Close()
	var buff []byte = make([]byte, 1024)
	var readBuffer []byte

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
			var rawCommand []byte = readBuffer[:n]
			readBuffer = readBuffer[n:]

			var result resp.RESPType
			var writeCommand bool
			switch cmd.Com {

			case "REPLCONF":
				// this used by slave to connect to master
				// cant be used by master to get ack, as handleConn function is not called for masterConn. (only client connections)
				result = client.Server.ReplconfHandler(cmd)
			case "PSYNC": // Master/Slave replication Commands
				result = client.Server.MasterServerCommandHandler(cmd)
			case "INFO": // Server Related Commands
				result = client.Server.ServerCommandHandler(cmd)
			case "WATCH", "UNWATCH", "MULTI", "EXEC", "DISCARD": // Transaction Commands handled by Client
				result = client.TransactionHandler(cmd, s)
			default: // All other commands handled by (Command package)
				if client.inTransaction {
					client.queuedCmds.Push(cmd)
					result = resp.SimpleString("QUEUED")
				} else {
					result, writeCommand = command.ExecuteCommand(cmd, s)
					if client.Server.Role == "master" && writeCommand && result.Type() != "SimpleError" {
						client.Server.NotifySlavesOfWrite(rawCommand)
					}
				}

			}

			response := resp.Encode(result)
			(client.Conn).Write(response)

			if client.Server.Role == "master" && cmd.Com == "PSYNC" {
				fmt.Printf("Sending RDB file to slave: %s\n", client.Conn.RemoteAddr())
				client.Server.sendRDBFile(client)
				// slave added in sendRDBFile function after sending RDB file to slave.
				fmt.Println("Added slave: ", client.Conn.RemoteAddr())
			}
		}
	}
}

func HandleMasterConn(client *Client, s *store.Store) {
	defer (client.Conn).Close()
	var buff []byte = make([]byte, 1024)
	var readBuffer []byte

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

			switch cmd.Com {
			case "REPLCONF":
				// used by master to get ack from slave
				result := client.Server.ReplconfHandler(cmd)
				response := resp.Encode(result)
				(client.Conn).Write(response)
			default:
				_, _ = command.ExecuteCommand(cmd, s)
			}
		}
	}
}
