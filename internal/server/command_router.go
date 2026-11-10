package server

import (
	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

// Called from handleConn(). So handling Clients . Master included as Client for a Slave server
// Routes commands based on server type (master/slave) and command type (replication, server, transaction, etc.)
// ExecuteCommand executes only client side commands
func routeCommand(client *Client, s *store.Store, cmd *resp.Command) resp.RESPType {
	var result resp.RESPType

	switch cmd.Com {

	case "PSYNC", "WAIT":
		// Master/Slave replication Commands
		result = client.Server.MasterServerCommandHandler(cmd)
	case "REPLCONF":
		// this used by slave to connect to master
		// cant be used by master to get ack, as handleConn function is not called for masterConn. (only client connections)
		result = client.Server.ReplconfHandler(cmd)
	case "INFO":
		// Server Related Commands
		result = client.Server.ServerCommandHandler(cmd)
	case "WATCH", "UNWATCH", "MULTI", "EXEC", "DISCARD":
		// Transaction Commands handled by Client
		result = client.TransactionHandler(cmd, s)
	default: // All other commands handled by (Command package)
		if client.inTransaction {
			client.queuedCmds.Push(cmd)
			result = resp.SimpleString("QUEUED")
		} else {
			result = command.ExecuteCommand(cmd, s)
		}

	}

	return result
}
