package server

import (
	"fmt"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

func (s *Server) connectToMaster(masterAddr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", masterAddr)
	if err != nil {
		fmt.Println("Error connecting to master: ", err.Error())
		return nil, err
	}

	var pingCmd = &resp.Command{Com: "PING", Args: []string{}}
	responseStr, err1 := s.writeCommandToMaster(pingCmd, conn)
	if err1 != nil {
		return nil, err1
	}
	if responseStr != "+PONG\r\n" {
		return nil, fmt.Errorf("ERR expected PONG from master for PING, got: %s", responseStr)
	}
	fmt.Println(responseStr)

	var replconfCmd = &resp.Command{Com: "REPLCONF", Args: []string{"listening-port", fmt.Sprintf("%d", s.Port)}}
	responseStr, err1 = s.writeCommandToMaster(replconfCmd, conn)
	if err1 != nil {
		return nil, err1
	}
	if responseStr != "+OK\r\n" {
		return nil, fmt.Errorf("ERR expected OK from master for REPLCONF, got: %s", responseStr)
	}
	fmt.Println(responseStr)

	var replconfCmd2 = &resp.Command{Com: "REPLCONF", Args: []string{"capa", "psync2"}}
	responseStr, err1 = s.writeCommandToMaster(replconfCmd2, conn)
	if err1 != nil {
		return nil, err1
	}
	if responseStr != "+OK\r\n" {
		return nil, fmt.Errorf("ERR expected OK from master for REPLCONF, got: %s", responseStr)
	}
	fmt.Println(responseStr)

	var psyncCmd = &resp.Command{Com: "PSYNC", Args: []string{"?", "-1"}}
	responseStr, err1 = s.writeCommandToMaster(psyncCmd, conn)
	if err1 != nil {
		return nil, err1
	}
	fmt.Println(responseStr)

	return conn, nil
}

func (s *Server) writeCommandToMaster(cmd *resp.Command, masterConn net.Conn) (string, error) {
	cmdArray := resp.Array{resp.BulkString(cmd.Com)}
	for _, arg := range cmd.Args {
		cmdArray = append(cmdArray, resp.BulkString(arg))
	}
	response := resp.Encode(cmdArray)

	_, err := masterConn.Write(response)
	if err != nil {
		return "", fmt.Errorf("ERR Error writing to master: %v", err)
	}
	fmt.Printf("Sent command to master: %s\n", cmd.Com)

	var buf []byte = make([]byte, 1024)
	n, err1 := masterConn.Read(buf)
	if err1 != nil {
		return "", fmt.Errorf("ERR Error reading from master: %v", err1)
	}
	return string(buf[:n]), nil
}

func (s *Server) addSlaveReplicationInfo(masterAddr string) {
	parts := strings.Split(masterAddr, " ")

	s.replicationInfo["role"] = "slave"
	s.replicationInfo["master_host"] = parts[0]
	s.replicationInfo["master_port"] = parts[1]
}
