package server

import (
	"fmt"
	"net"
	"github.com/google/uuid"

	"github.com/codecrafters-io/redis-starter-go/internal/collections"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

// Client represents a single client connection.
// QueuedCmds and watchKeys are stored per client - For transactoin support ( MULTI, EXEC, WATCH, DISCARD )
// Mutex is redundant sice each client handled in a separate go routine. No concurrent access to client data.
type Client struct {
	Conn   net.Conn
	Server *Server
	Id     string

	WriteChannel chan []byte 
	// buffered channel to write data to client. 
	// Pubsub subscribe shoudnt block client. Hence this is implemented
	// Writing to connection is done by channel

	queuedCmds    *collections.Queue[*resp.Command] // for multi/ exec transactions
	watchedKeys   map[string]int64                  // for watch command. Stores the key and its version at the time of WATCH command
	inTransaction bool                              // true if MULTI command has been issued and EXEC or DISCARD is pending. false otherwise.
}

func NewClient(conn net.Conn, server *Server) *Client {
	client := &Client{
		Conn:   conn,
		Server: server,
		Id:     uuid.New().String(),

		queuedCmds:    collections.NewQueue[*resp.Command](),
		watchedKeys:   make(map[string]int64),
		inTransaction: false,
		WriteChannel: make(chan []byte, 1000),
	}

	go client.WriteLoop()
	return client
}

func (c *Client) WriteLoop() {
	for msg := range c.WriteChannel {
		_, err := c.Conn.Write(msg)
		if err != nil {
			fmt.Printf("Error writing to client %s: %s: %s\n", c.Id, err.Error(), string(msg))
			break
		}
	}
}

func (c *Client) Send(msg []byte) {
	c.WriteChannel <- msg
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

func (c *Client) InSubscribedMode() bool {
	return c.Server.Pubsub.InSubsciberMode[c.Id]
}
