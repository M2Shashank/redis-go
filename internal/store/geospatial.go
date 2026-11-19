package store

import (
	"fmt"
	"strconv"
)

const MIN_LONGITUDE float64 = -180.0
const MAX_LONGITUDE float64 = 180.0
const MIN_LATITUDE float64 = -85.05112878
const MAX_LATITUDE float64 = 85.05112878

const LONGITUTE_RANGE float64 = MAX_LONGITUDE - MIN_LONGITUDE
const LATITUDE_RANGE float64 = MAX_LATITUDE - MIN_LATITUDE

// General helpers

func EncodeScore(lon, lat float64) int64 {
	// Normalisation to range of [0, 2^26) .
	normalisedLon := (1 << 26) * (lon - MIN_LONGITUDE) / LONGITUTE_RANGE
	normalisedLat := (1 << 26) * (lat - MIN_LATITUDE) / LATITUDE_RANGE

	// Converted to 26 bit Integers
	var intLon int32 = int32(normalisedLon)
	var intLat int32 = int32(normalisedLat)

	// Bits are interleaved to form a 52 bit integer score.
	var score, mult int64 = 0, 1
	for i := range 26 {
		score += int64((intLat>>i)&1) * mult
		score += int64((intLon>>i)&1) * (mult << 1)
		mult <<= 2
	}

	return score
}

func DecodeScore(score int64) (float64, float64) {
	// Bits are de-interleaved to form 26 bit integers for lon and lat
	var intLon, intLat int32 = 0, 0
	for i := range 26 {
		intLat |= int32((score>>(2*i))&1) << i
		intLon |= int32((score>>(2*i+1))&1) << i
	}

	// de-normalisation to thier original range.
	min_lon := MIN_LONGITUDE + (float64(intLon)*LONGITUTE_RANGE)/(1<<26)
	max_lon := MIN_LONGITUDE + (float64(intLon + 1)*LONGITUTE_RANGE)/(1<<26)
	min_lat := MIN_LATITUDE + (float64(intLat)*LATITUDE_RANGE)/(1<<26)
	max_lat := MIN_LATITUDE + (float64(intLat + 1)*LATITUDE_RANGE)/(1<<26)

	return (min_lon + max_lon) / 2, (min_lat + max_lat) / 2
}

// --

func (s *Store) AddGeoLocation(key string, lon, lat float64, name string) (bool, error) {
	score := EncodeScore(lon, lat)
	fmt.Printf("Encoded Score for (%f, %f) is %d\n", lon, lat, score)
	return s.ZADD(key, float64(score), name)
}

func (s *Store) GetLonLat(key string, name string) (float64, float64, error) {
	score, err := s.ZSCORE(key, name)
	if err != nil {
		return 0, 0, err
	}

	scr, _ := strconv.ParseInt(score, 10, 64)
	lon, lat := DecodeScore(scr)
	return lon, lat, nil
}
