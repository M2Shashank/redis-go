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

		if cmd.Args[1] == "dir" {
			result = resp.Array([]resp.RESPType{
				resp.BulkString(cmd.Args[1]),
				resp.BulkString(m.dir),
			})
		} else if cmd.Args[1] == "dbfilename" {
			result = resp.Array([]resp.RESPType{
				resp.BulkString(cmd.Args[1]),
				resp.BulkString(m.dbfilename),
			})
		}

	default:
		result = resp.SimpleError("ERR unknown command '" + cmd.Com + "'")

	}
	return result
}
