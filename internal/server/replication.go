package server

//-- Uses server specific items

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ----------------------------------- helper structs
type SlaveInfo struct {
	// used by master to track the offset of each slave.
	// notes down the expected offset of a slave after each write command
	expectedOffset int64
}
type ChannelItem struct {
	// used by master for receiving ACKs from slaves. Stores the offset sent by the slave and the client object of the slave.
	//* Used only for GETACK command.
	offset int64
	client *Client
}

// --------------------------------------------..........

// ------------------------------ Main Replication Manager Structs
type MasterManager struct {
	Replicas      map[*Client]*SlaveInfo // used by master to track connected slaves
	ReplicasMutex sync.RWMutex

	//! stores the client and the offset it sent
	//! the offset its supposed to have is stored in the SlaveInfo struct
	AckChannel       chan *ChannelItem // used by master to receive ACKs from slaves.
	hasPendingWrites bool              // used by master to track if atleast one write command has been sent to slaves.

	//!! general replication info. Can be shown in INFO command
	// Fields Commonn to both
	role               string // role of the server (master or slave)
	master_replid      string // id of the self server
	master_repl_offset int64  // for master its the offset of the last command sent to slaves. Hardcoded as 0 for master for now
}

type SlaveManager struct {
	Master *Client // used by slaves to track their master

	//!! general replication info. Can be shown in INFO command
	role          string // role of the server (master or slave)
	master_replid string // id of the master server
	repl_offset   int64  // for slave its the offset after the last command received from master

	master_host string // host of the master server used by slave
	master_port int    // port of the master server used by slave
}

// ------------------------------------................

// --------------------------- Object Factory functions

func NewMasterManager() *MasterManager {
	return &MasterManager{
		Replicas:           make(map[*Client]*SlaveInfo),
		ReplicasMutex:      sync.RWMutex{},
		AckChannel:         make(chan *ChannelItem, 100),
		hasPendingWrites:   false,
		role:               "master",
		master_replid:      "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb",
		master_repl_offset: 0,
	}
}

func NewSlaveManager(masterAddr string) *SlaveManager {
	parts := strings.Split(masterAddr, " ")
	port, _ := strconv.Atoi(parts[1])

	return &SlaveManager{
		Master:      nil,
		role:        "slave",
		master_host: parts[0],
		master_port: port,
	}
}

// ----------------------------------................

func (m *MasterManager) Info() map[string]string {
	var info map[string]string = make(map[string]string)
	info["role"] = m.role

	//? Only of master for now
	info["master_replid"] = m.master_replid
	info["master_repl_offset"] = strconv.FormatInt(m.master_repl_offset, 10)
	return info
}

func (s *SlaveManager) Info() map[string]string {
	var info map[string]string = make(map[string]string)
	info["role"] = s.role

	//? nothing specific for slave for now
	return info
}

// ----------------------------------................

//! Slave Manager methods

func (s *SlaveManager) HandleGETACK() []string {
	var offset int64 = s.repl_offset
	result := []string{
		"REPLCONF",
		"ACK",
		strconv.FormatInt(offset, 10),
	}
	return result
}

func (s *SlaveManager) GetMasterAddr() string {
	masterHost := s.master_host
	masterPort := s.master_port
	masterAddr := fmt.Sprintf("%s:%d", masterHost, masterPort)
	return masterAddr
}

func (s *SlaveManager) WriteCommandToMaster(rawCommand []byte) error {
	_, err := s.Master.Conn.Write(rawCommand)
	if err != nil {
		return fmt.Errorf("ERR Error writing to master: %v", err)
	}
	return nil
}

func (s *SlaveManager) readResponseFromMaster() (string, error) {
	var response []byte
	var previous byte

	for {
		var current [1]byte

		if _, err := io.ReadFull(s.Master.Conn, current[:]); err != nil {
			return "", fmt.Errorf("error reading from master: %v", err)
		}

		response = append(response, current[0])

		if previous == '\r' && current[0] == '\n' {
			return string(response), nil
		}

		previous = current[0]
	}
}

func (s *SlaveManager) IncrementOffset(increment int64) {
	s.repl_offset += increment
}

func (s *SlaveManager) SetPSYNCFields(FULLRESYNCFields []string) {
	// master_repid and master_repl_offset
	s.master_replid = FULLRESYNCFields[1]
	master_repl_offset, _ := strconv.ParseInt(FULLRESYNCFields[2], 10, 64)
	s.repl_offset = master_repl_offset
}

//! Master Manager mathods

// synchronous for now
func (m *MasterManager) WriteCommandToReplicas(rawCommand []byte) {
	m.ReplicasMutex.RLock()
	defer m.ReplicasMutex.RUnlock()

	for slave, slaveStruct := range m.Replicas {
		fmt.Printf("Sending to replica %s: %s\n", slave.Conn.RemoteAddr(), string(rawCommand))
		n, err := slave.Conn.Write(rawCommand)
		if err != nil {
			fmt.Println("Error writing to slave: ", err.Error())
			continue
		}
		fmt.Println("Notified slave of write: ", string(rawCommand))

		slaveStruct.expectedOffset += int64(n)
	}
}

func (m *MasterManager) DrainAckChannel() {
	for {
		select {
		case <-m.AckChannel:
		default:
			return
		}
	}
}

func (m *MasterManager) AddSlave(client *Client) {
	m.ReplicasMutex.Lock()
	defer m.ReplicasMutex.Unlock()
	m.Replicas[client] = &SlaveInfo{
		expectedOffset: 0,
	}
}

func (m *MasterManager) HandlePSYNC() string {
	result := fmt.Sprintf("FULLRESYNC %s %d", m.master_replid, m.master_repl_offset)
	return result
}

func (m *MasterManager) HandleACK(offsetStr string, client *Client) {
	offset, _ := strconv.ParseInt(offsetStr, 10, 64)
	m.AckChannel <- &ChannelItem{
		offset: offset,
		client: client,
	}
}

func (m *MasterManager) HandleWait(numSlaves int, timeLimit int, rawCommand []byte) int {
	if m.hasPendingWrites == false {
		return len(m.Replicas)
	}

	// drain channel to remove any acks from previous WAIT commands
	m.DrainAckChannel()

	m.WriteCommandToReplicas(rawCommand)

	var numValidAcks int = 0

	timer := time.NewTimer(time.Duration(timeLimit) * time.Millisecond)
forloop:
	for {
		select {
		case <-m.AckChannel:
			// Increment counter if an ack is received from a slave irrespective of the offset value.
			numValidAcks++

		case <-timer.C:
			break forloop
		}
	}

	return numValidAcks
}
