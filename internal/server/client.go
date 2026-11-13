package server

import (
	"net"

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
	return client.Server.MasterCommandHandler(cmd)
}

// commands handled by a Server. But both master and slave
func (client *Client) ServerCommandHandler(cmd *resp.Command) resp.RESPType {
	return client.Server.ServerCommandHandler(cmd)
}

// Persistence related function. Delegated to PersistenceManager.

func (client *Client) PersistenceHandler(cmd *resp.Command) resp.RESPType {
	return client.Server.Persistence.PersistanceHandler(cmd)
}