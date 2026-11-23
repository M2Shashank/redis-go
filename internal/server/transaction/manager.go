package transaction

import (
	"github.com/codecrafters-io/redis-starter-go/internal/collections"
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

type TransactionManager struct {
	// for multi/ exec transactions
	queuedCmds    *collections.Queue[*resp.Command] 
	
	// for watch command. Stores the key and its version at the time of WATCH command
	watchedKeys   map[string]int64
	
	// true if MULTI command has been issued and EXEC or DISCARD is pending. false otherwise.
	inTransaction bool                              
}

func NewTransactionManager() *TransactionManager {
	return &TransactionManager{
		queuedCmds:    collections.NewQueue[*resp.Command](),
		watchedKeys:   make(map[string]int64),
		inTransaction: false,
	}
}

func (m *TransactionManager) GetQueuedCmds() []*resp.Command {
	queuedCommands := make([]*resp.Command, 0, m.queuedCmds.Size())
	for !m.queuedCmds.Empty() {
		topCmd := m.queuedCmds.Pop()
		queuedCommands = append(queuedCommands, topCmd)
	}
	return queuedCommands
}

func (m *TransactionManager) SetVersion(key string, version int64) {
	m.watchedKeys[key] = version
}

func (m *TransactionManager) ClearWatchedKeys() {
	m.watchedKeys = make(map[string]int64)
}

func (m *TransactionManager) ClearQueuedCmds() {
	m.queuedCmds.Clear()
}

// used in transaction.go
func (m *TransactionManager) ClearTransactionState() {
	m.inTransaction = false
	m.ClearWatchedKeys()
	m.ClearQueuedCmds()
}

func (m *TransactionManager) GetWatchedKeys() map[string]int64 {
	return m.watchedKeys
}

func (m *TransactionManager) IsInTransaction() bool {
	return m.inTransaction
}

func (m *TransactionManager) SetInTransaction(transactionState bool) {
	m.inTransaction = transactionState
}

func (m *TransactionManager) PushCmdToQueue(cmd *resp.Command) {
	m.queuedCmds.Push(cmd)
}