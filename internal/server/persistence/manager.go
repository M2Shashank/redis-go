package persistence

import (
	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

type Manager struct {
	dir        string
	dbfilename string
}

func NewPersistenceManager(dir string, dbfilename string) *Manager {
	return &Manager{
		dir:        dir,
		dbfilename: dbfilename,
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
			temp = append(temp, "no")
		case "appenddirname":
			temp = append(temp, "appendonlydir")
		case "appendfilename":
			temp = append(temp, "appendonly.aof")
		case "appendfsync":
			temp = append(temp, "everysec")
		}

		result = resp.ArrayToBulkRESP(temp)

	default:
		result = resp.SimpleError("ERR unknown command '" + cmd.Com + "'")

	}
	return result
}
