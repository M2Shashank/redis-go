package server

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
	persistence "github.com/codecrafters-io/redis-starter-go/internal/server/persistence"
)

type Server struct {
	Port      int
	Role      string
	DataStore *store.Store

	// Replication Managers. A server uses one of these based on its role.
	MasterManager *MasterManager
	SlaveManager  *SlaveManager

	// Persistence Manager, Persistence related fields
	Persistence *persistence.Manager
}

func NewServer(flags []string) *Server {
	portStr, replicaof := flags[0], flags[1]
	port, _ := strconv.Atoi(portStr)
	var role string = "master"
	if replicaof != "" {
		role = "slave"
	}
	var manager *persistence.Manager = persistence.NewPersistenceManager(flags[2:])

	var masterManager *MasterManager = nil
	var slaveManager *SlaveManager = nil
	if role == "master" {
		masterManager = NewMasterManager()
	} else {
		slaveManager = NewSlaveManager(replicaof)
	}

	dataStore := store.NewStore()
	// Load data from RDB file if it exists. If not, start with an empty store.
	var rdbFile []byte = manager.Rdb.Data
	dataStore.LoadRDBIntoStore(rdbFile) // Parses and Loads data from RDB file into store

	return &Server{
		Port:               port,
		DataStore:          dataStore,
		Role:               role,
		MasterManager:      masterManager,
		SlaveManager:       slaveManager,
		Persistence:        manager,
	}
}

//************* Main Function. Accepts Client Connections ****************
// implemented for Master and Slave.

func Start(addr string, flags []string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	fmt.Printf("Server listening on, %v\n", lis.Addr())

	server := NewServer(flags)

	// Connect to master and perform handshake and start the connection if slave
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
		go HandleMasterConn(server.GetSlaveMaster())
	}

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

		go HandleConn(client)
	}
}

func (server *Server) ServerCommandHandler(cmd *resp.Command) resp.RESPType {
	switch cmd.Com {
	case "INFO":
		return InfoHandler(server)
	case "REPLCONF":
		return server.ReplconfHandler(cmd, nil)

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

	case "GETACK": // sent by master to slave
		// handles the REPLCONF commands sent by slave to master. So for 1
		result = s.SlaveReplconfHandler(cmd, client) 

	default:
		// handles the REPLCONF commands sent by master to slave. So for 2, 3, 4
		result = s.MasterReplconfHandler(cmd, client)

	}
	return result
}

func InfoHandler(server *Server) resp.RESPType {
	var replInfo map[string]string
	if server.Role == "master" {
		replInfo = server.MasterManager.Info()
	} else {
		replInfo = server.SlaveManager.Info()
	}

	info := strings.Builder{}
	fmt.Fprintf(&info, "# Replication\r\n")
	for key, value := range replInfo {
		fmt.Fprintf(&info, "%s:%s\r\n", key, value)
	}
	return resp.BulkString(info.String())
}