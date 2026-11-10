package server

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

type SlaveInfo struct {
	expectedOffset int64
}
type ChannelItem struct {
	offset int64
	client *Client
}

type Server struct {
	Port      int
	Addr      string
	Role      string
	DataStore *store.Store

	//replication
	Master        *Client              // used by slaves
	Replicas      map[*Client]*SlaveInfo // used by master to track connected slaves
	ReplicasMutex sync.RWMutex

	//! stores the client and the offset it sent
	//! the offset its supposed to have is stored in the SlaveInfo struct
	AckChannel chan *ChannelItem // used by master to receive ACKs from slaves. 

	Replication ReplicationInfo
}

func NewServer(port int, addr string, dataStore *store.Store) *Server {
	return &Server{
		Port:          port,
		Addr:          addr,
		DataStore:     dataStore,
		Replication:   ReplicationInfo{},
		Replicas:      make(map[*Client]*SlaveInfo),
		ReplicasMutex: sync.RWMutex{},
		AckChannel:    make(chan *ChannelItem, 10),
		Role:          "master", // default role is master. If replicaof specified, it will be changed to slave in server.Start()
	}
}

//************* Main Function. Accepts Client Connections ****************
// implemented for Master and Slave.

func Start(port int, addr string, dataStore *store.Store, replicaof string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	fmt.Printf("Server listening on, %v\n", lis.Addr())
	server := NewServer(port, addr, dataStore)
	server.SetBasicReplicationInfo(replicaof)

	if server.Role == "slave" {
		// Connect to master and perform handshake
		masterAddr := strings.ReplaceAll(replicaof, " ", ":")
		masterClient, err := server.connectToMaster(masterAddr)
		if err != nil {
			fmt.Println("Error connecting to master: ", err.Error())
			os.Exit(1)
		}
		server.Master = masterClient // MasterConnection is set for a slave
		err = server.InitiateHandshake()
		if err != nil {
			fmt.Println("Error initiating handshake: ", err.Error())
			os.Exit(1)
		}
		fmt.Printf("Connected to master at %s and completed the Handshake\n", masterAddr)

		// Start handling commands from master to replicate the data.
		// This is a blocking call and will run in the background for the lifetime of the slave server.
		go HandleMasterConn(server.Master, dataStore)
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
	default:
		return resp.SimpleError("ERR unknown command '" + cmd.Com + "'")
	}
}

// ******************************************************************

// Both HandleConn and HandleMasterConn use this function to handle REPLCONF. Handle all REPLCONF commands here
//* 1. REPLCONF GETACK * --> Sent my master to get acknowledgement of the offset of a slave
//* 2. REPLCONF listening-port <PORT> --> sent by slave to master during Handshake. port on which the slave is listening on. Only used for logging. Not used for any other purpose
//* 3. REPLCONF CAPA PSYNC2 --> Sent by slave to master During handshake. To say that the slave supports the PSYNC2 protocol.
//* 4. REPLCONF ACK <offset> --> Sent by slave to master to acknowledge the offset of the master. Master uses this to track the offset of each slave.
//! Replica-to-master communication uses HandleMasterConn on the replica.
//! Master-side replica connections use HandleConn.
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
		if s.Role == "slave" { // received by slave
			var offset int64 = s.Replication.master_repl_offset
			result = resp.Array([]resp.RESPType{
				resp.BulkString("REPLCONF"),
				resp.BulkString("ACK"),
				resp.BulkString(strconv.FormatInt(offset, 10)),
			})
		} // returned result to HandleMasterConn function

	case "ACK": // sent by slave
		if s.Role == "master" { // received by master
			offset, _ := strconv.ParseInt(cmd.Args[1], 10, 64)
			s.AckChannel <- &ChannelItem{
				offset: offset,
				client: client,
			}
		}
		result = resp.SimpleString("IGNORE")
		// Ignored in the HandleConn function. Master does not send any response back to slave for REPLCONF ACK command.

	default:
		result = resp.SimpleString("OK")

	}

	return result
}

func InfoHandler(server *Server) resp.RESPType {
	var replInfo map[string]string = server.Replication.Info()

	info := strings.Builder{}
	fmt.Fprintf(&info, "# Replication\r\n")
	for key, value := range replInfo {
		fmt.Fprintf(&info, "%s:%s\r\n", key, value)
	}
	return resp.BulkString(info.String())
}
