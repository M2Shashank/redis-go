package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/codecrafters-io/redis-starter-go/internal/server"
)

var port int
var replicaof string

var dir, dbfilename string
var appendonly, appenddirname, appendfilename, appendfsync string

func init() {
	flag.IntVar(&port, "port", 6379, "Port to listen on")
	flag.IntVar(&port, "p", 6379, "Port to listen on (shorthand)")

	flag.StringVar(&replicaof, "replicaof", "", "Replicate from the specified master")

	dir, _ = filepath.Abs(".")
	flag.StringVar(&dir, "dir", dir, "Directory to store the RDB file")
	flag.StringVar(&dbfilename, "dbfilename", "dump.rdb", "RDB file name")
	flag.StringVar(&appendonly, "appendonly", "no", "Append only mode (yes/no)")
	flag.StringVar(&appenddirname, "appenddirname", "appendonlydir", "Append only file directory")
	flag.StringVar(&appendfilename, "appendfilename", "appendonly.aof", "Append only file name")
	flag.StringVar(&appendfsync, "appendfsync", "everysec", "Append only file sync mode")
}

func main() {
	flag.Parse()

	var flags []string = []string{fmt.Sprintf("%d", port), replicaof, dir, dbfilename, appendonly, appenddirname, appendfilename, appendfsync}

	var addr string = fmt.Sprintf("0.0.0.0:%d", port)
	server.Start(addr, flags)
}
