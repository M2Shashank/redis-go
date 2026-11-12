package main

import (
	"flag"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/internal/server"
)

var port int
var replicaof string

var dir, dbfilename string

func init() {
	flag.IntVar(&port, "port", 6379, "Port to listen on")
	flag.IntVar(&port, "p", 6379, "Port to listen on (shorthand)")

	flag.StringVar(&replicaof, "replicaof", "", "Replicate from the specified master")

	flag.StringVar(&dir, "dir", ".", "Directory to store the RDB file")
	flag.StringVar(&dbfilename, "dbfilename", "dump.rdb", "RDB file name")
}

func main() {
	flag.Parse()

	var flags []string = []string{fmt.Sprintf("%d", port), replicaof, dir, dbfilename}

	var addr string = fmt.Sprintf("0.0.0.0:%d", port)
	server.Start(addr, flags)
}
