package main

import (
	"fmt"
	"net"
	"os"

    "github.com/codecrafters-io/redis-starter-go/internal/server"
    "github.com/codecrafters-io/redis-starter-go/internal/store"
)

func main() {
	dataStore := store.NewStore()

	lis, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	fmt.Printf("Server listening on port 6379, %v\n", lis.Addr())

	for {
		conn, err := lis.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go server.HandleConn(conn, dataStore)
	}
}