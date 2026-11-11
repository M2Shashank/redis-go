package server

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

// connecting to master, performing handshake is server's responsibility.
// sending commands to master and reading responses from master is handled by SlaveManager
// The SlaveManager is also responsible for managing replication state, like tracking master's replication ID and its own offset

//  These functions are used by slave servers only.

func (s *Server) connectToMaster() error {
	masterAddr := s.ReplicationManager.(*SlaveManager).GetMasterAddr()
	conn, err := net.Dial("tcp", masterAddr)
	if err != nil {
		fmt.Println("Error connecting to master: ", err.Error())
		return err
	}

	masterClient := NewClient(conn, s)
	s.ReplicationManager.(*SlaveManager).Master = masterClient // MasterConnection is set for a slave

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
	s.ReplicationManager.(*SlaveManager).SetPSYNCFields(FULLRESYNCFields)

	rdbHeader, err := s.readResponseFromMaster()
	if err != nil {
		return err
	}

	rdbLength, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(rdbHeader, "$"), "\r\n"))
	if err != nil || rdbLength < 0 {
		return fmt.Errorf("invalid RDB length %q", rdbHeader)
	}
	rdb := make([]byte, rdbLength)
	if _, err := io.ReadFull(s.ReplicationManager.(*SlaveManager).Master.Conn, rdb); err != nil {
		return fmt.Errorf("error reading RDB: %v", err)
	}
	fmt.Println("Received RDB file from master, length:", rdbLength)

	return nil
}

// used by slave to handle commands from master
// Commands from clients to slave are handled in HandleConn function, which takes care of sending responses back
// reading for commands from master and executing them on slave. No response is sent back to master for these commands.
// reading and writing commands to master from slave -> handleMasterConn function
// reading and writing commands to slave from master -> handleConn function
func HandleMasterConn(client *Client, s *store.Store) {
	defer (client.Conn).Close()
	var buff []byte = make([]byte, 1024)
	var readBuffer []byte

	for {
		n, err := (client.Conn).Read(buff)
		if err != nil {
			fmt.Println("Error reading from connection: ", err.Error())
			break
		}
		fmt.Printf("Read %d bytes from connection: %s\n", n, string(buff[:n]))
		readBuffer = append(readBuffer, buff[:n]...)

		for {
			n, cmd := resp.ParseRESPCommand(readBuffer)
			if cmd == nil || n == 0 {
				break
			}
			readBuffer = readBuffer[n:]

			switch cmd.Com {
			case "REPLCONF": // getack
				// scmd sent by master to get ack from slave
				result := client.Server.ReplconfHandler(cmd, nil)
				response := resp.Encode(result)
				(client.Conn).Write(response)
			default:
				if cmd.IsWriteCommand() {
					_ = command.ExecuteCommand(cmd, s)
				}
			}

			client.Server.IncreamentOffset(n)
		}
	}
}


// ------------------- helper methods for replication 

func (s *Server) IncreamentOffset(n int) {
	s.ReplicationManager.(*SlaveManager).IncrementOffset(int64(n))
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

	err := s.ReplicationManager.(*SlaveManager).WriteCommandToMaster(response)
	if err != nil {
		return err
	}
	fmt.Printf("Sent command to master: %s\n", cmd.Com)
	return nil
}

func (s *Server) readResponseFromMaster() (string, error) {
	return s.ReplicationManager.(*SlaveManager).readResponseFromMaster()
}

func (s *Server) GetSlaveMaster() *Client {
	return s.ReplicationManager.(*SlaveManager).Master
}


// ------------------------------------------