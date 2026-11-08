package server

import (
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

type Server struct {
	Port      int
	Addr      string
	DataStore *store.Store

	Connections     map[*Client]struct{} // remote address to client mapping
	ServerMutex     sync.RWMutex         // to protect the Connections map
	replicationInfo map[string]string    // to store replication info like master host, id, etc.
}

func NewServer(port int, addr string, dataStore *store.Store) *Server {
	return &Server{
		Port:            port,
		Addr:            addr,
		DataStore:       dataStore,
		Connections:     make(map[*Client]struct{}),
		replicationInfo: make(map[string]string),
		ServerMutex:     sync.RWMutex{},
	}
}

//************* Main Function. Accepts Client Connections ****************

func Start(port int, addr string, dataStore *store.Store, role string) {
	server := NewServer(port, addr, dataStore)
	server.infoReplication(role)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	fmt.Printf("Server listening on, %v\n", lis.Addr())

	for {
		conn, err := lis.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		} else {
			fmt.Println("Accepted connection from ", conn.RemoteAddr())
		}

		client := NewClient(conn, server)
		server.addClient(client)

		go HandleConn(client, dataStore)
	}
}

// ******************************************************************

// ------ Server Methods to manage Server info ------

func (s *Server) infoReplication(role string) {
	s.replicationInfo["role"] = role
	if role == "master" {
		s.replicationInfo["master_replid"] = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
		s.replicationInfo["master_repl_offset"] = "0"
	}
}

// -------- Server Methods to manage clients ---------

func (s *Server) removeClient(client *Client) {
	s.ServerMutex.Lock()
	defer s.ServerMutex.Unlock()
	delete(s.Connections, client)
}

func (s *Server) addClient(client *Client) {
	s.ServerMutex.Lock()
	defer s.ServerMutex.Unlock()
	s.Connections[client] = struct{}{}
}

// ----------------------------------------------------
