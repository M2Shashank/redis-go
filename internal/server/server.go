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

type Server struct {
	Port      int
	Addr      string
	Role      string
	DataStore *store.Store

	//replication
	Master        *Client              // used by slaves
	Replicas      map[*Client]struct{} // used by master to track connected slaves
	ReplicasMutex sync.RWMutex

	Replication ReplicationInfo
}

func NewServer(port int, addr string, dataStore *store.Store) *Server {
	return &Server{
		Port:          port,
		Addr:          addr,
		DataStore:     dataStore,
		Replication:   ReplicationInfo{},
		Replicas:      make(map[*Client]struct{}),
		ReplicasMutex: sync.RWMutex{},
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

	if replicaof != "" {
		server.Role = "slave"

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

func (s *Server) ReplconfHandler(cmd *resp.Command) resp.RESPType {
	if len(cmd.Args) < 2 {
		return resp.SimpleError("ERR wrong number of arguments for 'REPLCONF' command")
	}

	var result resp.RESPType = resp.SimpleString("OK")

	switch cmd.Args[0] {

	case "GETACK":
		if s.Role == "slave" {
			var offset int64 = s.Replication.master_repl_offset
			result = resp.Array([]resp.RESPType{
				resp.BulkString("REPLCONF"),
				resp.BulkString("ACK"),
				resp.BulkString(strconv.FormatInt(offset, 10)),
			})
		}

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
