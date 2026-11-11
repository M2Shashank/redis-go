package server

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

type Server struct {
	Port      int
	Role      string
	DataStore *store.Store

	//? persistence related fields. Used by both
	dir        string
	dbfilename string

	// Replication Manager
	ReplicationManager ReplicationManager
}

func NewServer(port int, dataStore *store.Store, role string, replicaof string) *Server {
	var manager ReplicationManager
	if role == "master" {
		manager = NewMasterManager()
	} else {
		manager = NewSlaveManager(replicaof)
	}

	return &Server{
		Port:               port,
		DataStore:          dataStore,
		Role:               role,
		ReplicationManager: manager,
	}
}

//************* Main Function. Accepts Client Connections ****************
// implemented for Master and Slave.

func Start(port int, addr string, dataStore *store.Store, flags []string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	fmt.Printf("Server listening on, %v\n", lis.Addr())

	var role string
	replicaof, dir, dbfilename := flags[0], flags[1], flags[2]
	if replicaof != "" {
		role = "slave"
	} else {
		role = "master"
	}

	server := NewServer(port, dataStore, role, replicaof)
	server.dir = dir
	server.dbfilename = dbfilename

	if server.Role == "slave" {
		// Connect to master and perform handshake

		//use master host and port fields of the slavemanager to connect to master.
		err := server.connectToMaster()
		if err != nil {
			fmt.Println("Error connecting to master: ", err.Error())
			os.Exit(1)
		}
		// The master connection is stored in the SlaveManager's Master field.
		// access the master using the s.ReplicationManager.(*SlaveManager).Master field. This is the client connection to the master.

		err = server.InitiateHandshake()
		if err != nil {
			fmt.Println("Error initiating handshake: ", err.Error())
			os.Exit(1)
		}
		fmt.Printf("Connected to master and completed the Handshake\n")

		// Start handling commands from master to replicate the data.
		// This is a blocking call and will run in the background for the lifetime of the slave server.
		go HandleMasterConn(server.GetSlaveMaster(), dataStore)
	}

	// slave server will also accept connections from clients.
	// Run for both master and slave servers to accept connections from clients.
	for {
		conn, err := lis.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		} else {
			fmt.Println("Accepted connection from ", conn.RemoteAddr())
		}

		client := NewClient(conn, server)

		go HandleConn(client, dataStore)
	}
}

func (s *Server) ServerCommandHandler(cmd *resp.Command) resp.RESPType {
	switch cmd.Com {
	case "INFO":
		return InfoHandler(s)
	case "CONFIG":
		if len(cmd.Args) != 2 {
			return resp.SimpleError("ERR wrong number of arguments for 'CONFIG' command")
		}

		var typeOf string = cmd.Args[0]
		if typeOf != "GET" {
			return resp.SimpleError("ERR unknown subcommand '" + typeOf + "' for 'CONFIG' command")
		}

		var result resp.RESPType
		if cmd.Args[1] == "dir" {
			result = resp.Array([]resp.RESPType{
				resp.BulkString(cmd.Args[1]),
				resp.BulkString(s.dir),
			})
		} else if cmd.Args[1] == "dbfilename" {
			result = resp.Array([]resp.RESPType{
				resp.BulkString(cmd.Args[1]),
				resp.BulkString(s.dbfilename),
			})
		}
		return result

	default:
		return resp.SimpleError("ERR unknown command '" + cmd.Com + "'")
	}
}

// ******************************************************************

// Both HandleConn and HandleMasterConn use this function to handle REPLCONF. Handle all REPLCONF commands here
// * 1. REPLCONF GETACK * --> Sent my master to get acknowledgement of the offset of a slave
// * 2. REPLCONF listening-port <PORT> --> sent by slave to master during Handshake. port on which the slave is listening on. Only used for logging. Not used for any other purpose
// * 3. REPLCONF CAPA PSYNC2 --> Sent by slave to master During handshake. To say that the slave supports the PSYNC2 protocol.
// * 4. REPLCONF ACK <offset> --> Sent by slave to master to acknowledge the offset of the master. Master uses this to track the offset of each slave.
// ! Replica-to-master communication uses HandleMasterConn on the replica.
// ! Master-side replica connections use HandleConn.
// if server is master -> valid commands are 2, 3, 4 (sent by slave). No response is sent back for 4. Response for 2, 3 is "OK". [handler called by HandleConn function for 2, 3, 4]
// if server is slave -> valid commands are only 1 (sent by master). command 4 is sent as response as bulkarray [handler called by HandleMasterConn function for 1]
func (s *Server) ReplconfHandler(cmd *resp.Command, client *Client) resp.RESPType {
	// client is the one which sent the REPLCONF command.
	if len(cmd.Args) < 2 {
		return resp.SimpleError("ERR wrong number of arguments for 'REPLCONF' command")
	}

	var result resp.RESPType = resp.SimpleString("OK")

	switch strings.ToUpper(cmd.Args[0]) {

	case "GETACK": // sent by master
		// slave handles the GETACK command and sends the current offset of the slave to the master.
		result = s.ReplicationManager.(*SlaveManager).HandleGETACK()
		// returned result to HandleMasterConn function

	case "ACK": // sent by slave
		// master handles the ACK and adds it to the AckChannel.
		s.ReplicationManager.(*MasterManager).HandleACK(cmd.Args[1], client) // No response is sent back to slave for this command.
		result = resp.SimpleString("IGNORE")
		// Ignored in the HandleConn function. Master does not send any response back to slave for REPLCONF ACK command.

	default:
		result = resp.SimpleString("OK")

	}

	return result
}

func InfoHandler(server *Server) resp.RESPType {
	var replInfo map[string]string = server.GetInfo()

	info := strings.Builder{}
	fmt.Fprintf(&info, "# Replication\r\n")
	for key, value := range replInfo {
		fmt.Fprintf(&info, "%s:%s\r\n", key, value)
	}
	return resp.BulkString(info.String())
}

// ***************************************************
// ***************************************************
// ***************************************************
// TO be moved to replica and master files

func (s *Server) GetInfo() map[string]string {
	result := s.ReplicationManager.Info()
	return result
}

// -----------------------------------------


// ------------------------------------------

