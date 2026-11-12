package server

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

// connecting to master, performing handshake is server's responsibility.
// sending commands to master and reading responses from master is handled by SlaveManager
// The SlaveManager is also responsible for managing replication state, like tracking master's replication ID and its own offset

//  These functions are used by slave servers only.

func (s *Server) connectToMaster() error {
	masterAddr := s.SlaveManager.GetMasterAddr()
	conn, err := net.Dial("tcp", masterAddr)
	if err != nil {
		fmt.Println("Error connecting to master: ", err.Error())
		return err
	}

	masterClient := NewClient(conn, s)
	s.SlaveManager.Master = masterClient // MasterConnection is set for a slave

	return nil
}

func (s *Server) InitiateHandshake() error {
	var pingCmd = &resp.Command{
		Com:  "PING",
		Args: []string{},
	}
	var pingCmdIdealResponse = "+PONG\r\n"
	_, err := s.SendAndCheckCommandToMaster(pingCmd, pingCmdIdealResponse)
	if err != nil {
		return err
	}

	var replconfCmd = &resp.Command{
		Com:  "REPLCONF",
		Args: []string{"listening-port", fmt.Sprintf("%d", s.Port)},
	}
	var replconfCmdIdealResponse = "+OK\r\n"
	_, err = s.SendAndCheckCommandToMaster(replconfCmd, replconfCmdIdealResponse)
	if err != nil {
		return err
	}

	var replconfCmd2 = &resp.Command{
		Com:  "REPLCONF",
		Args: []string{"capa", "psync2"},
	}
	_, err = s.SendAndCheckCommandToMaster(replconfCmd2, replconfCmdIdealResponse)
	if err != nil {
		return err
	}

	var psyncCmd = &resp.Command{
		Com:  "PSYNC",
		Args: []string{"?", "-1"},
	}
	resp, err := s.SendAndCheckCommandToMaster(psyncCmd, "+FULLRESYNC")
	if err != nil {
		return err
	}	

	FULLRESYNCFields := strings.Fields(resp)
	s.SlaveManager.SetPSYNCFields(FULLRESYNCFields)

	rdbHeader, err := s.readResponseFromMaster()
	if err != nil {
		return err
	}

	rdbLength, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(rdbHeader, "$"), "\r\n"))
	if err != nil || rdbLength < 0 {
		return fmt.Errorf("invalid RDB length %q", rdbHeader)
	}
	rdb := make([]byte, rdbLength)
	if _, err := io.ReadFull(s.SlaveManager.Master.Conn, rdb); err != nil {
		return fmt.Errorf("error reading RDB: %v", err)
	}
	fmt.Println("Received RDB file from master, length:", rdbLength)

	return nil
}



// ------------------- helper methods for replication 

func (s *Server) IncreamentOffset(n int) {
	s.SlaveManager.IncrementOffset(int64(n))
}

// -------------------------- Handshake helpers 

func (s *Server) SendAndCheckCommandToMaster(cmd *resp.Command, expectedResponse string) (string, error) {
	err := s.writeCommandToMaster(cmd)
	if err != nil {
		return "", err
	}
	responseStr, err := s.readResponseFromMaster()
	if err != nil {
		return "", err
	}
	if !strings.Contains(responseStr, expectedResponse) {
		return "", fmt.Errorf("ERR expected %s from master, got: %s", expectedResponse, responseStr)
	}

	fmt.Println(responseStr)
	return responseStr, nil
}

func (s *Server) writeCommandToMaster(cmd *resp.Command) error {
	cmdArray := resp.CommandToRESP(cmd)
	response := resp.Encode(cmdArray)

	err := s.SlaveManager.WriteCommandToMaster(response)
	if err != nil {
		return err
	}
	fmt.Printf("Sent command to master: %s\n", cmd.Com)
	return nil
}

func (s *Server) readResponseFromMaster() (string, error) {
	return s.SlaveManager.readResponseFromMaster()
}

func (s *Server) GetSlaveMaster() *Client {
	return s.SlaveManager.Master
}


// ------------------------------------------



// Replconf commands include:
// * 1. REPLCONF GETACK * --> Sent my master to get acknowledgement of the offset of a slave
// * 2. REPLCONF listening-port <PORT> --> sent by slave to master during Handshake. port on which the slave is listening on. Only used for logging. Not used for any other purpose
// * 3. REPLCONF CAPA PSYNC2 --> Sent by slave to master During handshake. To say that the slave supports the PSYNC2 protocol.
// * 4. REPLCONF ACK <offset> --> Sent by slave to master to acknowledge the offset of the master. Master uses this to track the offset of each slave.
// ! Replica-to-master communication uses HandleMasterConn on the replica.
// ! Master-side replica connections use HandleConn.
// if server is master -> valid commands are 2, 3, 4 (sent by slave). No response is sent back for 4. Response for 2, 3 is "OK". [handler called by HandleConn function for 2, 3, 4]
// if server is slave -> valid commands are only 1 (sent by master). command 4 is sent as response as bulkarray [handler called by HandleMasterConn function for 1]
// This function is for handling REPLCONF commands sent by master to slave. So for 1
func (s *Server) SlaveReplconfHandler(cmd *resp.Command, client *Client) resp.RESPType {
	var result resp.RESPType = resp.SimpleString("OK")

	switch strings.ToUpper(cmd.Args[0]) {

	case "GETACK": // sent by master
		// slave handles the GETACK command and sends the current offset of the slave to the master.
		result = s.SlaveManager.HandleGETACK()
		// returned result to HandleMasterConn function

	default:
		// ideally should not reach here
		result = resp.SimpleString("OK")

	}

	return result
}