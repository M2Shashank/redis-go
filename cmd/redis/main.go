package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/internal/server"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

var port int
var dataStore *store.Store

func init() {
	dataStore = store.NewStore()
	flag.IntVar(&port, "port", 6379, "Port to listen on")
	flag.IntVar(&port, "p", 6379, "Port to listen on (shorthand)")
}

func main() {
	flag.Parse()

	var addr string = fmt.Sprintf("0.0.0.0:%d", port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	fmt.Printf("Server listening on, %v\n", lis.Addr())

	for {
		conn, err := lis.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		} else {
			fmt.Println("Accepted connection from ", conn.RemoteAddr())
		}

		client := server.NewClient(&conn)
		go server.HandleConn(client, dataStore)
	}
}
