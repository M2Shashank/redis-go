package server

import (
	"encoding/hex"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

var emptyRDBFileHex string = "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"

func (s *Server) MasterServerCommandHandler(cmd *resp.Command) resp.RESPType {
	if s.Role != "master" {
		return resp.SimpleError("ERR unknown command '" + cmd.Com + "'")
	}

	var result resp.RESPType
	switch cmd.Com {

	case "PSYNC":
		masterId := s.Replication.master_replid
		masterOffset := s.Replication.master_repl_offset
		str := fmt.Sprintf("FULLRESYNC %s %d", masterId, masterOffset)
		result = resp.SimpleString(str)
	case "WAIT":
		if len(cmd.Args) != 2 {
			return resp.SimpleError("ERR wrong number of arguments for 'WAIT' command")
		}

		result = resp.Integer(0)
		// For now, we are not implementing WAIT command fully. So returning 0 for now.
	default:
		result = resp.SimpleError("ERR unknown command '" + cmd.Com + "'")

	}
	return result
}

func (s *Server) AddSlave(client *Client) {
	s.ReplicasMutex.Lock()
	defer s.ReplicasMutex.Unlock()
	s.Replicas[client] = struct{}{}
}

func (s *Server) SendRDBFile(client *Client) {
	rdbFile, _ := hex.DecodeString(emptyRDBFileHex)
	str := fmt.Sprintf("$%d\r\n", len(rdbFile))
	(client.Conn).Write([]byte(str))
	(client.Conn).Write(rdbFile)
}

// Synchronous for now
func (s *Server) WriteCommandToSlaves(rawCommand []byte) {
	s.ReplicasMutex.RLock()
	defer s.ReplicasMutex.RUnlock()

	for slave := range s.Replicas {
		fmt.Printf("Sending to replica %s: %s\n", slave.Conn.RemoteAddr(), string(rawCommand))
		_, err := slave.Conn.Write(rawCommand)
		if err != nil {
			fmt.Println("Error writing to slave: ", err.Error())
		}
		fmt.Println("Notified slave of write: ", string(rawCommand))
	}
}
