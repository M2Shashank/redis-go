package server

import (
	"github.com/codecrafters-io/redis-starter-go/internal/cmd_router"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"

	Error "github.com/codecrafters-io/redis-starter-go/internal/errors"
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

	return resp.SimpleError(Error.ERR_INVALID.Error())
}

func (c *Client) HandleWatch(cmd *resp.Command) resp.RESPType {
	if c.inTransaction {
		return resp.SimpleError(Error.ERR_WATCH_INSIDE_MULTI.Error())
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
		return resp.SimpleError(Error.ERR_MULTI_NESTED.Error())
	}

	c.inTransaction = true
	return resp.SimpleString("OK")
}

func (c *Client) HandleExec() resp.RESPType {
	if !c.inTransaction {
		return resp.SimpleError(Error.ERR_EXEC_WITHOUT_MULTI.Error())
	}

	if !c.isValidTransaction() {
		c.clearTransactionState()
		return resp.Array(nil)
	}

	var results resp.Array = make([]resp.RESPType, 0, c.queuedCmds.Size())
	for !c.queuedCmds.Empty() {
		topCmd := c.queuedCmds.Pop()
		result := command.StoreCommandHandler(topCmd, c.Server.DataStore)
		results = append(results, result)
	}
	c.clearTransactionState()
	return results
}

func (c *Client) HandleDiscard() resp.RESPType {
	if !c.inTransaction {
		return resp.SimpleError(Error.ERR_DISCARD_WITHOUT_MULTI.Error())
	}

	c.clearTransactionState()
	return resp.SimpleString("OK")
}
