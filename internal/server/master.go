package server

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

var emptyRDBFileHex string = "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"

func (s *Server) AddSlave(client *Client) {
	s.MasterManager.AddSlave(client)
}

func (s *Server) SendRDBFile(client *Client) {
	rdbFile, _ := hex.DecodeString(emptyRDBFileHex)
	str := fmt.Sprintf("$%d\r\n", len(rdbFile))
	(client.Conn).Write([]byte(str))
	(client.Conn).Write(rdbFile)
}

func (s *Server) syncReplicasOfWrite(rawCommand []byte, com string) {
	if com != "PING" && com != "REPLCONF" && com != "PSYNC" && com != "INFO" && com != "WAIT" {
		s.MasterManager.hasPendingWrites = true
		s.MasterManager.WriteCommandToReplicas(rawCommand)
	}
}

// Replconf commands include:
// * 1. REPLCONF GETACK * --> Sent my master to get acknowledgement of the offset of a slave
// * 2. REPLCONF listening-port <PORT> --> sent by slave to master during Handshake. port on which the slave is listening on. Only used for logging. Not used for any other purpose
// * 3. REPLCONF CAPA PSYNC2 --> Sent by slave to master During handshake. To say that the slave supports the PSYNC2 protocol.
// * 4. REPLCONF ACK <offset> --> Sent by slave to master to acknowledge the offset of the master. Master uses this to track the offset of each slave.
// ! Replica-to-master communication uses HandleMasterConn on the replica.
// ! Master-side replica connections use HandleConn.
// if server is master -> valid commands are 2, 3, 4 (sent by slave). No response is sent back for 4. Response for 2, 3 is "OK". [handler called by HandleConn function for 2, 3, 4]
// if server is slave -> valid commands are only 1 (sent by master). command 4 is sent as response as bulkarray [handler called by HandleMasterConn function for 1]
// This function is for handling REPLCONF commands sent by slave to master. So for 2, 3, 4
func (s *Server) MasterReplconfHandler(cmd *resp.Command, client *Client) resp.RESPType {
	var result resp.RESPType = resp.SimpleString("OK")

	switch strings.ToUpper(cmd.Args[0]) {

	case "ACK": // sent by slave
		// master handles the ACK and adds it to the AckChannel.
		s.MasterManager.HandleACK(cmd.Args[1], client) // No response is sent back to slave for this command.
		result = resp.SimpleString("IGNORE")
		// Ignored in the HandleConn function. Master does not send any response back to slave for REPLCONF ACK command.

	default:
		result = resp.SimpleString("OK")

	}

	return result
}

// Master Commands Handler

func (server *Server) MasterCommandHandler(cmd *resp.Command) resp.RESPType {
	if server.Role != "master" {
		return resp.SimpleError("ERR unknown command '" + cmd.Com + "'")
	}

	var result resp.RESPType
	switch cmd.Com {

	case "PSYNC":
		temp := server.MasterManager.HandlePSYNC()
		result = resp.SimpleString(temp)

	case "WAIT":
		if len(cmd.Args) != 2 {
			return resp.SimpleError("ERR wrong number of arguments for 'WAIT' command")
		}
		numSlaves, err1 := strconv.Atoi(cmd.Args[0])
		timeLimit, err2 := strconv.Atoi(cmd.Args[1])
		if err1 != nil || err2 != nil {
			return resp.SimpleError("ERR invalid arguments for 'WAIT' command")
		}

		ReplConfGetackCmd := &resp.Command{
			Com:  "REPLCONF",
			Args: []string{"GETACK", "*"},
		}
		temp := resp.CommandToRESP(ReplConfGetackCmd)
		rawCommand := resp.Encode(temp)

		var noOfValidAcks int = server.MasterManager.HandleWait(numSlaves, timeLimit, rawCommand)
		result = resp.Integer(noOfValidAcks)

	default:
		result = resp.SimpleError("ERR unknown command '" + cmd.Com + "'")

	}
	return result
}
