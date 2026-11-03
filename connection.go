package main

import (
	"fmt"
	"net"
)

func handleConn(conn net.Conn, store *Store) {
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
			n, cmd := parseRESPCommand(readBuffer)
			if cmd == nil || n == 0 {
				break
			}

			ExecuteCommand(conn, cmd, store)
			readBuffer = readBuffer[n:]
		}
	}

}