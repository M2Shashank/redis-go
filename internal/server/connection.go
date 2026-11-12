package server

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/command"
)

// used by both master and slave to handle client connections
// slave can send commands to master thorugh this function (as master is also a client for slave)
// Master cannot send commands to slave through this function, as handleConn function is not called for masterConn. (only client connections)
func HandleClientConn(client *Client) {
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
			result = routeCommand(client, cmd)

			if client.Server.Role == "master" && result.Type() != "SimpleError" && cmd.Com != "PING" && cmd.Com != "REPLCONF" && cmd.Com != "PSYNC" && cmd.Com != "INFO" && cmd.Com != "WAIT" {
				client.Server.MasterManager.hasPendingWrites = true
				client.Server.WriteCommandToSlaves(rawCommand)
			}

			if cmd.Com == "REPLCONF" && cmd.Args[0] == "ACK" {
				continue
			}

			response := resp.Encode(result)
			(client.Conn).Write(response)

			// routeCommand for PSYNC acknowledges repl_id and offset. Now the RDB file is sent
			if client.Server.Role == "master" && cmd.Com == "PSYNC" {
				fmt.Printf("Sending RDB file to slave: %s\n", client.Conn.RemoteAddr())
				client.Server.SendRDBFile(client)
				client.Server.AddSlave(client) // slave added in master's list
				fmt.Println("Added slave: ", client.Conn.RemoteAddr())
			}
		}
	}
}



// Called from handleConn(). So handling Clients . Master included as Client for a Slave server
// Routes commands based on server type (master/slave) and command type (replication, server, transaction, etc.)
// ExecuteCommand executes only client side commands
func routeCommand(client *Client, cmd *resp.Command) resp.RESPType {
	var result resp.RESPType

	switch cmd.Com {

	case "PSYNC", "WAIT":
		// master implements these commands.
		result = client.Server.MasterServerCommandHandler(cmd)
	case "INFO", "CONFIG", "REPLCONF":
		//  Related Commands
		result = client.Server.ServerCommandHandler(cmd)
	case "WATCH", "UNWATCH", "MULTI", "EXEC", "DISCARD":
		// Transaction Commands handled by Client
		result = client.TransactionHandler(cmd)
	default: // All other commands handled by (Command package)
		if client.inTransaction {
			client.queuedCmds.Push(cmd)
			result = resp.SimpleString("QUEUED")
		} else {
			result = command.ExecuteCommand(cmd, client.Server.DataStore)
		}

	}

	return result
}