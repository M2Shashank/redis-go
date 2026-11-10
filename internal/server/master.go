package server

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/command"
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
		numSlaves, err1 := strconv.Atoi(cmd.Args[0])
		timeLimit, err2 := strconv.Atoi(cmd.Args[1])
		if err1 != nil || err2 != nil {
			return resp.SimpleError("ERR invalid arguments for 'WAIT' command")
		}

		result = s.HandleWait(numSlaves, timeLimit)

	default:
		result = resp.SimpleError("ERR unknown command '" + cmd.Com + "'")

	}
	return result
}

func (s *Server) HandleWait(numSlaves int, timeLimit int) resp.RESPType {
	ReplConfGetackCmd := &resp.Command{
		Com:  "REPLCONF",
		Args: []string{"GETACK", "*"},
	}
	temp := command.CommandToRESP(ReplConfGetackCmd)
	rawCommand := resp.Encode(temp)
	s.WriteCommandToSlaves(rawCommand, false)

	timer := time.NewTimer(time.Duration(timeLimit) * time.Millisecond)
	defer timer.Stop()

	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	for {
		if s.countAcknowledgedSlaves() >= numSlaves {
			return resp.Integer(s.countAcknowledgedSlaves())
		}

		select {
		case <-ticker.C:
		case <-timer.C:
			return resp.Integer(s.countAcknowledgedSlaves())
		}
	}
}

func (s *Server) AddSlave(client *Client) {
	s.ReplicasMutex.Lock()
	defer s.ReplicasMutex.Unlock()
	s.Replicas[client] = &SlaveInfo{
		expectedOffset: 0,
	}
}

func (s *Server) SendRDBFile(client *Client) {
	rdbFile, _ := hex.DecodeString(emptyRDBFileHex)
	str := fmt.Sprintf("$%d\r\n", len(rdbFile))
	(client.Conn).Write([]byte(str))
	(client.Conn).Write(rdbFile)
}

// Synchronous for now
func (s *Server) WriteCommandToSlaves(rawCommand []byte, incrOffset bool) {
	s.ReplicasMutex.Lock()
	defer s.ReplicasMutex.Unlock()

	for slave, slaveStruct := range s.Replicas {
		fmt.Printf("Sending to replica %s: %s\n", slave.Conn.RemoteAddr(), string(rawCommand))
		n, err := slave.Conn.Write(rawCommand)
		if err != nil {
			fmt.Println("Error writing to slave: ", err.Error())
			continue
		}
		fmt.Println("Notified slave of write: ", string(rawCommand))

		if incrOffset {
			slaveStruct.expectedOffset += int64(n)
		}
	}
}

func (s *Server) countAcknowledgedSlaves() int {
	s.ReplicasMutex.RLock()
	defer s.ReplicasMutex.RUnlock()

	count := 0
	for _, slaveInfo := range s.Replicas {
		if slaveInfo.acknowledgedOffset >= slaveInfo.expectedOffset {
			count++
		}
	}
	return count
}
