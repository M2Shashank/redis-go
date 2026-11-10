// * Handles all helper functions related to replication ------------------
package server

import (
	"strconv"
	"strings"
)

// Replication of a Server stored as a struct. stored Inside the server

// Store's replication info of server.
type ReplicationInfo struct {
	// Fields Commonn to both
	role          string // role of the server (master or slave)
	master_replid string // id of the master server

	master_repl_offset int64
	//! for master its the offset of the last command sent to slaves, for slave its the offset of the last command received from master
	//!hardcoded as 0 for master. Calculated for slave

	// Fields specific to slave
	master_host string // host of the master server used by slave
	master_port int    // port of the master server used by slave

	// Fields specific to master
}

func (r *ReplicationInfo) Info() map[string]string {
	var info map[string]string = make(map[string]string)
	info["role"] = r.role

	if r.role == "slave" {
		//? Nothing for now.
		// info["master_host"] = r.master_host
		// info["master_port"] = strconv.Itoa(r.master_port)
	} else {
		//? Only of master for now
		info["master_replid"] = r.master_replid
		info["master_repl_offset"] = strconv.FormatInt(r.master_repl_offset, 10)
	}
	return info
}

// ______________________________________________________________

func (s *Server) SetBasicReplicationInfo(masterAddr string) {
	if masterAddr == "" {
		s.Replication.role = "master"
		s.Replication.master_replid = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
		s.Replication.master_repl_offset = 0
	} else {
		parts := strings.Split(masterAddr, " ")

		s.Replication.role = "slave"
		s.Replication.master_host = parts[0]
		s.Replication.master_port, _ = strconv.Atoi(parts[1])
	}
}
