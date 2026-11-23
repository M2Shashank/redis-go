package handlers

import (
	"fmt"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
	"github.com/codecrafters-io/redis-starter-go/internal/store"

	Error "github.com/codecrafters-io/redis-starter-go/internal/errors"
)

func Geospacial(cmd *resp.Command, dataStore *store.Store) resp.RESPType {
	switch cmd.Com {
	case "GEOADD":
		var noOfEntries int = 0
		for i := 1; i < len(cmd.Args); i += 3 {
			longitude, err1 := strconv.ParseFloat(cmd.Args[i], 64)
			latitude, err2 := strconv.ParseFloat(cmd.Args[i+1], 64)
			ErrorStr := Error.ERR_INVALID_GEO_COORDINATES(cmd.Args[i], cmd.Args[i+1]).Error()

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
		result := resp.Array{}
		for i := 1; i < len(cmd.Args); i++ {
			lon, lat, err := dataStore.GetLonLat(cmd.Args[0], cmd.Args[i])

			var curResult resp.RESPType
			if err != nil {
				curResult = resp.Array(nil)
			} else {
				curResult = resp.Array{
					resp.BulkString(fmt.Sprintf("%f", lon)),
					resp.BulkString(fmt.Sprintf("%f", lat)),
				}
			}
			result = append(result, curResult)
		}
		return result

	case "GEODIST":
		distance, err := dataStore.GetDistanceMeters(cmd.Args[0], cmd.Args[1], cmd.Args[2])
		if err != nil {
			return resp.SimpleError(err.Error())
		}

		return resp.BulkString(fmt.Sprintf("%f", distance))

	case "GEOSEARCH":
		// Only implementing for FROMLONLAT , BYRADIUS, m
		lon, err1 := strconv.ParseFloat(cmd.Args[2], 64)
		lat, err2 := strconv.ParseFloat(cmd.Args[3], 64)
		radius, err3 := strconv.ParseFloat(cmd.Args[5], 64)
		if err1 != nil || (lon < -180 || lon > 180) {
			return resp.SimpleError(Error.ERR_INVALID_LON(cmd.Args[2]).Error())
		}
		if err2 != nil || (lat < -85.05112878 || lat > +85.05112878) {
			return resp.SimpleError(Error.ERR_INVALID_LAT(cmd.Args[3]).Error())
		}
		if err3 != nil || radius < 0 {
			return resp.SimpleError(Error.ERR_NAN.Error())
		}

		list := dataStore.GeoSearch(cmd.Args[0], lon, lat, radius)
		result := resp.ArrayToBulkRESP(list)
		return result
	}

	return resp.SimpleError(Error.ERR_INVALID.Error())
}
