package persistence

import (
	"path/filepath"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

type Manager struct {
	dir        string
	dbfilename string
	appendonly   string
	appenddirname string
	appendfilename string
	appendfsync  string
}

func NewPersistenceManager(flags []string) *Manager {
	if flags[2] == "yes" {
		// create a new appendonly directory if it doesn't exist
		appendonlyDir := filepath.Join(flags[0], flags[3])
		CreateAppendOnlyDir(appendonlyDir)

	}

	return &Manager{
		dir:        flags[0],
		dbfilename: flags[1],
		appendonly:   flags[2],
		appenddirname: flags[3],
		appendfilename: flags[4],
		appendfsync:  flags[5],
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
			temp = append(temp, m.dir)
		case "dbfilename":
			temp = append(temp, m.dbfilename)
		case "appendonly":
			temp = append(temp, m.appendonly)
		case "appenddirname":
			temp = append(temp, m.appenddirname)
		case "appendfilename":
			temp = append(temp, m.appendfilename)
		case "appendfsync":
			temp = append(temp, m.appendfsync)
		}

		result = resp.ArrayToBulkRESP(temp)

	default:
		result = resp.SimpleError("ERR unknown command '" + cmd.Com + "'")

	}
	return result
}
