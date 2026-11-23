package Errors

import (
	"errors"
)

// This file contains all the errors that can be returned by the server.

func ERR_WRONGTYPE(command string) error {
	return errors.New("ERR WrongType Operation. NOT A " + command)
}

//! STRING related errors
var ERR_NAN = errors.New("ERR value is not an integer or out of range")
var ERR_NAN_FLOAT = errors.New("ERR value is not a valid float")

//! ZSET related errors
var ERR_MEMBER_NOT_FOUND = errors.New("ERR one or both members do not exist in the sorted set")
var ERR_NOT_FOUND = errors.New("ERR Key or field does not exist")

//! Geospatial related errors
func ERR_INVALID_GEO_COORDINATES(long, lat string) error {
	return errors.New("ERR invalid longitude, latitude pair " + long + ", " + lat)
}
func ERR_INVALID_LAT(lat string) error {
	return errors.New("ERR invalid latitude " + lat + ". Must be between -90 and 90")
}
func ERR_INVALID_LON(lon string) error {
	return errors.New("ERR invalid longitude " + lon + ". Must be between -180 and 180")
}