package server

import (
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func HandleConn(client *Client, s *store.Store) {
	defer (*client.Conn).Close()
	var buff []byte = make([]byte, 1024)
	var readBuffer []byte

	for {
		n, err := (*client.Conn).Read(buff)
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			break
		}

		readBuffer = append(readBuffer, buff[:n]...)

		for {
			n, cmd := resp.ParseRESPCommand(readBuffer)
			if cmd == nil || n == 0 {
				break
			}
			readBuffer = readBuffer[n:]

			if cmd.Com == "WATCH" {
				var result resp.RESPType
				if client.inTransaction {
					result = resp.SimpleError("ERR WATCH inside MULTI is not allowed")
				} else {
					for _, key := range cmd.Args {
						client.setWatchKey(key, s)
					}
					result = resp.SimpleString("OK")
				}
				response := resp.Encode(result)
				(*client.Conn).Write(response)
				continue
			}

			if cmd.Com == "UNWATCH" {
				client.clearWatchedKeys()
				response := resp.Encode(resp.SimpleString("OK"))
				(*client.Conn).Write(response)
				continue
			}

			if cmd.Com == "MULTI" {
				var result resp.RESPType
				if client.inTransaction {
					result = resp.SimpleError("ERR MULTI calls can not be nested")
				} else {
					result = resp.SimpleString("OK")
					client.inTransaction = true
				}
				response := resp.Encode(result)
				(*client.Conn).Write(response)
				continue
			}

			if cmd.Com == "EXEC" {
				if !client.inTransaction {
					result := resp.SimpleError("ERR EXEC without MULTI")
					response := resp.Encode(result)
					(*client.Conn).Write(response)
					continue
				}

				var response []byte
				if !client.isValidTransaction(s) {
					response = resp.Encode(resp.Array(nil))
				} else {
					var results resp.Array = make([]resp.RESPType, 0, client.queuedCmds.Size())
					for {
						if client.queuedCmds.Empty() {
							break
						}
						c := client.queuedCmds.Pop()
						result := command.ExecuteCommand(c, s)
						results = append(results, result)
					}
					response = resp.Encode(results)
				}
				(*client.Conn).Write(response)
				client.inTransaction = false
				client.clearWatchedKeys()
				client.clearQueuedCmds()
				continue
			}

			if cmd.Com == "DISCARD" {
				var result resp.RESPType
				if client.inTransaction {
					result = resp.SimpleString("OK")
				} else {
					result = resp.SimpleError("ERR DISCARD without MULTI")
				}
				response := resp.Encode(result)
				(*client.Conn).Write(response)
				client.inTransaction = false
				client.clearWatchedKeys()
				client.clearQueuedCmds()
				continue
			}

			if client.inTransaction {
				client.queuedCmds.Push(cmd)
				response := resp.Encode(resp.SimpleString("QUEUED"))
				(*client.Conn).Write(response)
				continue
			}

			result := command.ExecuteCommand(cmd, s)
			response := resp.Encode(result)

			(*client.Conn).Write(response)
		}
	}
}
