package server

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/internal/collections"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/command"
)

// Client represents a single client connection.
// QueuedCmds and watchKeys are stored per client - For transactoin support ( MULTI, EXEC, WATCH, DISCARD )
// Mutex is redundant sice each client handled in a separate go routine. No concurrent access to client data.
type Client struct {
	Conn net.Conn
	Server *Server

	queuedCmds *collections.Queue[*resp.Command] // for multi/ exec transactions
	watchedKeys map[string]int64 // for watch command. Stores the key and its version at the time of WATCH command
	inTransaction bool // true if MULTI command has been issued and EXEC or DISCARD is pending. false otherwise.
}

func NewClient(conn net.Conn, server *Server) *Client {
	return &Client{
		Conn: conn,
		Server: server,
		queuedCmds: collections.NewQueue[*resp.Command](),
		watchedKeys: make(map[string]int64),
		inTransaction: false,
	}
}

func (c *Client) setWatchKey(key string) {
	var version int64 = c.Server.DataStore.GetKeyVersion(key)
	c.watchedKeys[key] = version
}

func (c *Client) clearWatchedKeys() {
	c.watchedKeys = make(map[string]int64)
}

func (c *Client) clearQueuedCmds() {
	c.queuedCmds.Clear()
}

func (c *Client) isValidTransaction() bool {
	for key, version := range c.watchedKeys {
		currentVersion := c.Server.DataStore.GetKeyVersion(key)
		if currentVersion != version {
			return false
		}
	}
	return true
}

// used in transaction.go
func (c *Client) clearTransactionState() {
	c.inTransaction = false
	c.clearWatchedKeys()
	c.clearQueuedCmds()
}



// functions related to handling connections

// Called from handleConn(). So handling Clients . Master included as Client for a Slave server
// Routes commands based on server type (master/slave) and command type (replication, server, transaction, etc.)
// Command Package's ExecuteCommand executes only client side commands
func (client *Client) ExecuteCommand(cmd *resp.Command) resp.RESPType {
	var result resp.RESPType

	switch cmd.Com {

	case "PSYNC", "WAIT":
		// master implements these commands.
		result = client.MasterCommandHandler(cmd)
	case "INFO", "CONFIG", "REPLCONF":
		// Server Related Commands. Both master and slave
		result = client.ServerCommandHandler(cmd)
	case "WATCH", "UNWATCH", "MULTI", "EXEC", "DISCARD":
		// Transaction Commands handled by Client
		result = client.TransactionHandler(cmd)
	default: // All other commands handled by (Command package)
		if client.inTransaction {
			client.queuedCmds.Push(cmd)
			result = resp.SimpleString("QUEUED")
		} else {
			result = command.StoreCommandHandler(cmd, client.Server.DataStore)
		}

	}

	return result
}

// commands handled by a Server. But both master and slave
func (client *Client) ServerCommandHandler(cmd *resp.Command) resp.RESPType {
	var server *Server = client.Server

	switch cmd.Com {
	case "INFO":
		return InfoHandler(server)
	case "REPLCONF":
		return server.ReplconfHandler(cmd, nil)
	case "CONFIG": // Persistance related command . Will be moved to Persistance Manager later
		if len(cmd.Args) != 2 {
			return resp.SimpleError("ERR wrong number of arguments for 'CONFIG' command")
		}

		var typeOf string = cmd.Args[0]
		if typeOf != "GET" {
			return resp.SimpleError("ERR unknown subcommand '" + typeOf + "' for 'CONFIG' command")
		}

		var result resp.RESPType
		if cmd.Args[1] == "dir" {
			result = resp.Array([]resp.RESPType{
				resp.BulkString(cmd.Args[1]),
				resp.BulkString(server.dir),
			})
		} else if cmd.Args[1] == "dbfilename" {
			result = resp.Array([]resp.RESPType{
				resp.BulkString(cmd.Args[1]),
				resp.BulkString(server.dbfilename),
			})
		}
		return result

	default:
		return resp.SimpleError("ERR unknown command '" + cmd.Com + "'")
	}
}