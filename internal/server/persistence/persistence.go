package persistence

import (
	"path/filepath"
)

type Manager struct {
	Dir            string
	Dbfilename     string
	Appendonly     string
	Appenddirname  string
	Appendfilename string
	Appendfsync    string

	Rdb *RDBManager
	Aof *AOFManager
}

func NewPersistenceManager(flags []string) (*Manager, []byte) {
	dir, dbfilename, appendonly, appenddirname, appendfilename, appendfsync := flags[0], flags[1], flags[2], flags[3], flags[4], flags[5]

	var aofManager *AOFManager = nil
	var fileData []byte = nil
	if appendonly == "yes" {
		// create a new appendonly directory if it doesn't exist
		appendonlyDir := filepath.Join(dir, appenddirname)

		aofManager, fileData = NewAOFManager(appendonlyDir, appendfilename)
	}

	var rdb *RDBManager = NewRDBManager(dir, dbfilename)

	return &Manager{
		Dir:            dir,
		Dbfilename:     dbfilename,
		Appendonly:     appendonly,
		Appenddirname:  appenddirname,
		Appendfilename: appendfilename,
		Appendfsync:    appendfsync,

		Aof: aofManager,
		Rdb: rdb,
	}, fileData
}

func (m *Manager) HandleConfigGET(cmdType string) []string {
	var temp []string = []string{cmdType}

	switch cmdType {
	case "dir":
		temp = append(temp, m.Dir)
	case "dbfilename":
		temp = append(temp, m.Dbfilename)
	case "appendonly":
		temp = append(temp, m.Appendonly)
	case "appenddirname":
		temp = append(temp, m.Appenddirname)
	case "appendfilename":
		temp = append(temp, m.Appendfilename)
	case "appendfsync":
		temp = append(temp, m.Appendfsync)
	}

	return temp
}

func (m *Manager) FlushWriteCommandToAOF(cmd []byte) {
	if m.Aof == nil {
		return
	}

	m.Aof.WriteCommand(cmd, m.Appendfsync)
}