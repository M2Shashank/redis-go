package server

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func InfoHandler(cmd *resp.Command, server *Server) resp.RESPType {
	info := strings.Builder{}
	fmt.Fprintf(&info, "# Replication\r\n")
	for key, value := range server.replicationInfo {
		fmt.Fprintf(&info, "%s:%s\r\n", key, value)
	}
	return resp.BulkString(info.String())
}
