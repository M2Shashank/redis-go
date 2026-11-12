package server

import (
	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

// ------ Transaction Commands Implementation -------
// ---- ( MULTI, EXEC, WATCH, DISCARD, UNWATCH ) ----
// implemented in server bacause they need access to Client struct fields and methods

func (client *Client) TransactionHandler(cmd *resp.Command) resp.RESPType {
	switch cmd.Com {
	case "WATCH":
		return client.HandleWatch(cmd)
	case "UNWATCH":
		return client.HandleUnwatch()
	case "MULTI":
		return client.HandleMulti()
	case "EXEC":
		return client.HandleExec()
	case "DISCARD":
		return client.HandleDiscard()
	}
	return resp.SimpleError("ERR unknown command")
}

func (c *Client) HandleWatch(cmd *resp.Command) resp.RESPType {
	if c.inTransaction {
		return resp.SimpleError("ERR WATCH inside MULTI is not allowed")
	}

	for _, key := range cmd.Args {
		c.setWatchKey(key)
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

func (c *Client) HandleExec() resp.RESPType {
	if !c.inTransaction {
		return resp.SimpleError("ERR EXEC without MULTI")
	}

	if !c.isValidTransaction() {
		c.clearTransactionState()
		return resp.Array(nil)
	}

	var results resp.Array = make([]resp.RESPType, 0, c.queuedCmds.Size())
	for !c.queuedCmds.Empty() {
		topCmd := c.queuedCmds.Pop()
		result := command.ExecuteCommand(topCmd, c.Server.DataStore)
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
