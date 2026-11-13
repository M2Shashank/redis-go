package persistence

import (
	"path/filepath"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
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

func NewPersistenceManager(flags []string) *Manager {
	dir, dbfilename, appendonly, appenddirname, appendfilename, appendfsync := flags[0], flags[1], flags[2], flags[3], flags[4], flags[5]

	var aof *AOFManager = nil
	if appendonly == "yes" {
		// create a new appendonly directory if it doesn't exist
		appendonlyDir := filepath.Join(dir, appenddirname)
		filename := appendfilename + ".1.incr.aof"

		aof = NewAOFManager(appendonlyDir, filename)
	}

	var rdb *RDBManager = NewRDBManager(dir, dbfilename)

	return &Manager{
		Dir:            dir,
		Dbfilename:     dbfilename,
		Appendonly:     appendonly,
		Appenddirname:  appenddirname,
		Appendfilename: appendfilename,
		Appendfsync:    appendfsync,

		Aof: aof,
		Rdb: rdb,
	}
}

func (m *Manager) PersistanceHandler(cmd *resp.Command) resp.RESPType {
	var result resp.RESPType
	switch cmd.Com {

	case "CONFIG":
		if len(cmd.Args) != 2 {
			return resp.SimpleError("ERR wrong number of arguments for 'CONFIG' command")
		}

		var typeOf string = cmd.Args[0]
		if typeOf != "GET" {
			return resp.SimpleError("ERR unknown subcommand '" + typeOf + "' for 'CONFIG' command")
		}

		var temp []string = []string{cmd.Args[1]}

		switch cmd.Args[1] {
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

		result = resp.ArrayToBulkRESP(temp)

	default:
		result = resp.SimpleError("ERR unknown command '" + cmd.Com + "'")

	}
	return result
}
