package server

import (
	"net"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/internal/collections"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
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



func (client *Client) MasterCommandHandler(cmd *resp.Command) resp.RESPType {
	var server *Server = client.Server
	var server_repl__manager *MasterManager = server.MasterManager
	
	if server.Role != "master" {
		return resp.SimpleError("ERR unknown command '" + cmd.Com + "'")
	}

	var result resp.RESPType
	switch cmd.Com {

	case "PSYNC":
		server.MasterManager.HandlePSYNC()

	case "WAIT":
		if len(cmd.Args) != 2 {
			return resp.SimpleError("ERR wrong number of arguments for 'WAIT' command")
		}
		numSlaves, err1 := strconv.Atoi(cmd.Args[0])
		timeLimit, err2 := strconv.Atoi(cmd.Args[1])
		if err1 != nil || err2 != nil {
			return resp.SimpleError("ERR invalid arguments for 'WAIT' command")
		}

		var noOfValidAcks int = server_repl__manager.HandleWait(numSlaves, timeLimit)
		result = resp.Integer(noOfValidAcks)

	default:
		result = resp.SimpleError("ERR unknown command '" + cmd.Com + "'")

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

	default:
		return resp.SimpleError("ERR unknown command '" + cmd.Com + "'")
	}
}

// Persistence related function. Delegated to PersistenceManager.

func (client *Client) PersistenceHandler(cmd *resp.Command) resp.RESPType {
	return client.Server.Persistence.PersistanceHandler(cmd)
}