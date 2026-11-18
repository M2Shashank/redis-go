package command

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func geoHandler(cmd *resp.Command, dataStore *store.Store) resp.RESPType {
	switch cmd.Com {
	case "GEOADD":
		if len(cmd.Args) < 4 || len(cmd.Args)%3 != 1 {
			return resp.SimpleError("ERR wrong number of arguments for 'geoadd' command")
		}

		for i := 1; i < len(cmd.Args); i += 3 {
			longitude, err1 := strconv.ParseFloat(cmd.Args[i], 64)
			latitude, err2 := strconv.ParseFloat(cmd.Args[i+1], 64)
			ErrorStr := fmt.Sprintf("ERR invalid longitude, latitude pair %s, %s", cmd.Args[i], cmd.Args[i+1])

			if err1 != nil || (longitude < -180 || longitude > 180) {
				return resp.SimpleError(ErrorStr)
			}
			if err2 != nil || (latitude < -85.05112878 || latitude > +85.05112878) {
				return resp.SimpleError(ErrorStr)
			}
		}

		newEntries := (len(cmd.Args) - 1) / 3
		return resp.Integer(newEntries)
	default:
		return resp.SimpleError("ERR Invalid or Unsupported Command")
	}
}
