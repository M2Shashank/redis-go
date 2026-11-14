package server

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

// used by both master and slave to handle client connections
// slave can send commands to master thorugh this function (as master is also a client for slave)
// Master cannot send commands to slave through this function, as handleConn function is not called for masterConn. (only client connections)
func HandleConn(client *Client) {
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

			var result resp.RESPType = client.ExecuteCommand(cmd)
			var response []byte = resp.Encode(result)

			// If the command is a write command and the server is a master, send it to the replicas
			if client.Server.Role == "master" && result.Type() != "SimpleError" {
				client.Server.syncReplicasOfWrite(rawCommand, cmd.Com)
			}

			if cmd.Com == "REPLCONF" && cmd.Args[0] == "ACK" {
				// Dont write back to slave
				continue
			}

			if result.Type() != "SimpleError" && cmd.IsWriteCommand() {
				client.Server.Persistence.FlushWriteCommandToAOF(rawCommand)
			}

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

// used by slave to handle commands from master
// Commands from clients to slave are handled in HandleConn function, which takes care of sending responses back
// reading for commands from master and executing them on slave. No response is sent back to master for these commands.
// reading and writing commands to master from slave -> handleMasterConn function
// reading and writing commands to slave from master -> handleConn function
func HandleMasterConn(client *Client) {
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
			case "REPLCONF": // getack
				// cmd sent by master to get ack from slave
				result := client.Server.ReplconfHandler(cmd, nil)
				response := resp.Encode(result)
				(client.Conn).Write(response)
			default:
				if cmd.IsWriteCommand() {
					_ = client.ExecuteCommand(cmd)
				}
			}

			client.Server.SlaveManager.IncrementOffset(int64(n))
		}
	}
}

// Called from handleConn(). So handling Clients . Master included as Client for a Slave server
// Routes commands based on server type (master/slave) and command type (replication, server, transaction, etc.)
// Command Package's ExecuteCommand executes only client side commands
func (client *Client) ExecuteCommand(cmd *resp.Command) resp.RESPType {
	var result resp.RESPType

	switch cmd.Com {

	case "PSYNC", "WAIT":
		// PSYNC sent by slave, WAIT sent by client.
		// Both sent to master. .: Handled by MasterCommandHandler
		result = client.Server.MasterCommandHandler(cmd)
	case "INFO", "REPLCONF":
		// INFO sent by client. REPLCONF - getack sent by master, remaining sent by slave.
		// Server Related Commands. Both master and slave
		result = client.Server.ServerCommandHandler(cmd)
	default:
		// client handeled commands
		if client.inSubscribedMode {
			perm := client.Pubsub.ValidateCommand(cmd)
			if !perm {
				result = resp.SimpleError("ERR Can't execute '" + cmd.Com + "': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context")
			}
			break
		}

		switch cmd.Com {
		case "CONFIG":
			// Config commands handled by PersistenceManager.
			result = client.PersistenceHandler(cmd)
		case "WATCH", "UNWATCH", "MULTI", "EXEC", "DISCARD":
			// Transaction Commands handled by Client
			result = client.TransactionHandler(cmd)
		case "SUBSCRIBE":
			var mode bool
			result, mode = client.PubSubHandler(cmd)
			client.inSubscribedMode = mode
		default: // All other commands handled by (Command package)
			if client.inTransaction {
				client.queuedCmds.Push(cmd)
				result = resp.SimpleString("QUEUED")
			} else {
				result = command.StoreCommandHandler(cmd, client.Server.DataStore)
			}
		}

	}

	return result
}

// Persistence related function. Delegated to PersistenceManager.

func (client *Client) PersistenceHandler(cmd *resp.Command) resp.RESPType {
	return client.Server.Persistence.PersistanceHandler(cmd)
}

// PubSub related function. Delegated to PubSubManager.

func (client *Client) PubSubHandler(cmd *resp.Command) (resp.RESPType, bool) {
	return client.Pubsub.HandleCommand(cmd), true
}
