package server

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/internal/store"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/command"
)

func HandleConn(conn net.Conn, store *store.Store) {
	defer conn.Close()
	var buff []byte = make([]byte, 1024)
	var readBuffer []byte

	for {
		n, err := conn.Read(buff)
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

			result := command.ExecuteCommand(cmd, store)
			response := resp.Encode(result)

			conn.Write(response)
		}
	}
}