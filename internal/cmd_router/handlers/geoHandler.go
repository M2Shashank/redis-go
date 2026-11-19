package handlers

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

func Geospacial(cmd *resp.Command, dataStore *store.Store) resp.RESPType {
	switch cmd.Com {
	case "GEOADD":
		if len(cmd.Args) < 4 || len(cmd.Args)%3 != 1 {
			return resp.SimpleError("ERR wrong number of arguments for 'geoadd' command")
		}

		var noOfEntries int = 0
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

			added, err := dataStore.AddGeoLocation(cmd.Args[i-1], longitude, latitude, cmd.Args[i+2])
			if err != nil {
				return resp.SimpleError(err.Error())
			}
			if added {
				noOfEntries++
			}
		}
		return resp.Integer(noOfEntries)

	case "GEOPOS":
		if len(cmd.Args) < 2 {
			return resp.SimpleError("ERR wrong number of arguments for 'geopos' command")
		}

		result := resp.Array{}
		for i := 1; i < len(cmd.Args); i++ {
			lon, lat, err := dataStore.GetLonLat(cmd.Args[0], cmd.Args[i])

			var curResult resp.RESPType
			if err != nil {
				curResult = append(result, resp.Array(nil))
			} else {
				curResult = resp.Array{
					resp.BulkString(fmt.Sprintf("%f", lon)),
					resp.BulkString(fmt.Sprintf("%f", lat)),
				}
			}
			result = append(result, curResult)
		}
		return result

	default:
		return resp.SimpleError("ERR Invalid or Unsupported Command")
	}
}
