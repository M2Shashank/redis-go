package server

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/internal/collections"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

type Client struct {
	Conn *net.Conn

	QueuedCmds *collections.Queue[*resp.Command]
	watchedKeys map[string]bool
	inTransaction bool
}

func NewClient(conn *net.Conn) *Client {
	return &Client{
		Conn: conn,
		QueuedCmds: collections.NewQueue[*resp.Command](),
		watchedKeys: make(map[string]bool),
		inTransaction: false,
	}
}