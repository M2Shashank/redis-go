package server

import (
	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

// ------ Transaction Commands Implementation -------
// ---- ( MULTI, EXEC, WATCH, DISCARD, UNWATCH ) ----
// implemented in server bacause they need access to Client struct fields and methods

func (c *Client) TransactionHandler(cmd *resp.Command, s *store.Store) resp.RESPType {
	switch cmd.Com {
	case "WATCH":
		return c.HandleWatch(cmd, s)
	case "UNWATCH":
		return c.HandleUnwatch()
	case "MULTI":
		return c.HandleMulti()
	case "EXEC":
		return c.HandleExec(s)
	case "DISCARD":
		return c.HandleDiscard()
	}
	return resp.SimpleError("ERR unknown command")
}

func (c *Client) HandleWatch(cmd *resp.Command, s *store.Store) resp.RESPType {
	if c.inTransaction {
		return resp.SimpleError("ERR WATCH inside MULTI is not allowed")
	}

	for _, key := range cmd.Args {
		c.setWatchKey(key, s)
	}
	return resp.SimpleString("OK")
}

func (c *Client) HandleUnwatch() resp.RESPType {
	c.clearWatchedKeys()
	return resp.SimpleString("OK")
}

func (c *Client) HandleMulti() resp.RESPType {
	if c.inTransaction {
		return resp.SimpleError("ERR MULTI calls can not be nested")
	}

	c.inTransaction = true
	return resp.SimpleString("OK")
}

func (c *Client) HandleExec(s *store.Store) resp.RESPType {
	if !c.inTransaction {
		return resp.SimpleError("ERR EXEC without MULTI")
	}

	if !c.isValidTransaction(s) {
		c.clearTransactionState()
		return resp.Array(nil)
	}

	var results resp.Array = make([]resp.RESPType, 0, c.queuedCmds.Size())
	for !c.queuedCmds.Empty() {
		topCmd := c.queuedCmds.Pop()
		result := command.ExecuteCommand(topCmd, s)
		results = append(results, result)
	}
	c.clearTransactionState()
	return results
}

func (c *Client) HandleDiscard() resp.RESPType {
	if !c.inTransaction {
		return resp.SimpleError("ERR DISCARD without MULTI")
	}

	c.clearTransactionState()
	return resp.SimpleString("OK")
}
