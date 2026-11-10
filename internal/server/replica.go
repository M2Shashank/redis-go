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

func (s *Server) connectToMaster(masterAddr string) (*Client, error) {
	conn, err := net.Dial("tcp", masterAddr)
	if err != nil {
		fmt.Println("Error connecting to master: ", err.Error())
		return nil, err
	}

	masterClient := NewClient(conn, s)
	return masterClient, nil
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

	fullResyncFields := strings.Fields(resp)
	s.replicationInfo["ack_repl_offset"] = fullResyncFields[2]
	s.replicationInfo["master_replid"] = fullResyncFields[1]

	rdbHeader, err := s.readResponseFromMaster(s.Master.Conn)
	if err != nil {
		return err
	}

	rdbLength, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(rdbHeader, "$"), "\r\n"))
	if err != nil || rdbLength < 0 {
		return fmt.Errorf("invalid RDB length %q", rdbHeader)
	}
	rdb := make([]byte, rdbLength)
	if _, err := io.ReadFull(s.Master.Conn, rdb); err != nil {
		return fmt.Errorf("error reading RDB: %v", err)
	}
	fmt.Println("Received RDB file from master, length:", rdbLength)

	return nil
}

func (s *Server) SendAndCheckCommandToMaster(cmd *resp.Command, expectedResponse string) (string, error) {
	err := s.writeCommandToMaster(cmd, s.Master.Conn)
	if err != nil {
		return "", err
	}
	responseStr, err := s.readResponseFromMaster(s.Master.Conn)
	if err != nil {
		return "", err
	}
	if !strings.Contains(responseStr, expectedResponse) {
		return "", fmt.Errorf("ERR expected %s from master, got: %s", expectedResponse, responseStr)
	}

	fmt.Println(responseStr)
	return responseStr, nil
}

func (s *Server) writeCommandToMaster(cmd *resp.Command, masterConn net.Conn) error {
	cmdArray := command.CommandToRESP(cmd)
	response := resp.Encode(cmdArray)

	_, err := masterConn.Write(response)
	if err != nil {
		return fmt.Errorf("ERR Error writing to master: %v", err)
	}
	fmt.Printf("Sent command to master: %s\n", cmd.Com)
	return nil
}

func (s *Server) readResponseFromMaster(masterConn net.Conn) (string, error) {
	var response []byte
	var previous byte

	for {
		var current [1]byte

		if _, err := io.ReadFull(masterConn, current[:]); err != nil {
			return "", fmt.Errorf("error reading from master: %v", err)
		}

		response = append(response, current[0])

		if previous == '\r' && current[0] == '\n' {
			return string(response), nil
		}

		previous = current[0]
	}
}

func (s *Server) addSlaveReplicationInfo(masterAddr string) {
	parts := strings.Split(masterAddr, " ")

	s.replicationInfo["master_host"] = parts[0]
	s.replicationInfo["master_port"] = parts[1]
}

func (s *Server) increamentOffset(n int) {
	curOffsetState := s.replicationInfo["ack_repl_offset"]
	offset, _ := strconv.ParseInt(curOffsetState, 10, 64)
	offset += int64(n)
	s.replicationInfo["ack_repl_offset"] = strconv.FormatInt(offset, 10)
}

// used by slave to handle commands from master
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
			case "REPLCONF":
				// used by master to get ack from slave
				result := client.Server.ReplconfHandler(cmd)
				response := resp.Encode(result)
				(client.Conn).Write(response)
			default:
				_, _ = command.ExecuteCommand(cmd, s)
			}

			client.Server.increamentOffset(n)
		}
	}
}
