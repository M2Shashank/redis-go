package server

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/internal/collections"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

// Client represents a single client connection.
// QueuedCmds and watchKeys are stored per client - For transactoin support ( MULTI, EXEC, WATCH, DISCARD )
// Mutex is redundant sice each client handled in a separate go routine. No concurrent access to client data.
type Client struct {
	Conn *net.Conn

	queuedCmds *collections.Queue[*resp.Command]
	watchedKeys map[string]int64
	inTransaction bool
}

func NewClient(conn *net.Conn) *Client {
	return &Client{
		Conn: conn,
		queuedCmds: collections.NewQueue[*resp.Command](),
		watchedKeys: make(map[string]int64),
		inTransaction: false,
	}
}

func (c *Client) setWatchKey(key string, s *store.Store) {
	var version int64 = s.GetKeyVersion(key)
	c.watchedKeys[key] = version
}

func (c *Client) clearWatchedKeys() {
	c.watchedKeys = make(map[string]int64)
}

func (c *Client) isValidTransaction(s *store.Store) bool {
	for key, version := range c.watchedKeys {
		currentVersion := s.GetKeyVersion(key)
		if currentVersion != version {
			return false
		}
	}
	return true
}