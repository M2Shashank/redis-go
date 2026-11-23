package server

import (
	"fmt"
	"net"
	"github.com/google/uuid"

	"github.com/codecrafters-io/redis-starter-go/internal/server/transaction"
)

// Client represents a single client connection.
// QueuedCmds and watchKeys are stored per client - For transactoin support ( MULTI, EXEC, WATCH, DISCARD )
// Mutex is redundant sice each client handled in a separate go routine. No concurrent access to client data.
type Client struct {
	Conn   net.Conn
	Server *Server
	Id     string
	Username string

	WriteChannel chan []byte 
	// buffered channel to write data to client. 
	// Pubsub subscribe shoudnt block client. Hence this is implemented
	// Writing to connection is done by channel

	// manages transaction state for the client
	TransactionManager *transaction.TransactionManager

	// Checks if a client is authenticated or not
	Authenticated bool
}

func NewClient(conn net.Conn, server *Server) *Client {
	authenticated := server.AuthVerdict("default")

	client := &Client{
		Conn:   conn,
		Server: server,
		Id:     uuid.New().String(),

		TransactionManager: transaction.NewTransactionManager(),

		WriteChannel: make(chan []byte, 1000),

		Username: "default",
		Authenticated: authenticated,
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

func (c *Client) InSubscribedMode() bool {
	return c.Server.Pubsub.InSubsciberMode[c.Id]
}
