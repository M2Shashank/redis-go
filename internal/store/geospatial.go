package store

import (
	"fmt"
	"math"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/internal/store/storage"

	Error "github.com/codecrafters-io/redis-starter-go/internal/errors"
)

const EARTH_RADIUS_METERS float64 = 6372797.560856

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
	max_lon := MIN_LONGITUDE + (float64(intLon+1)*LONGITUTE_RANGE)/(1<<26)
	min_lat := MIN_LATITUDE + (float64(intLat)*LATITUDE_RANGE)/(1<<26)
	max_lat := MIN_LATITUDE + (float64(intLat+1)*LATITUDE_RANGE)/(1<<26)

	return (min_lon + max_lon) / 2, (min_lat + max_lat) / 2
}

func HaversineDistance(lon1, lat1, lon2, lat2 float64) float64 {
	lonRad1, latRad1 := lon1*math.Pi/180, lat1*math.Pi/180
	lonRad2, latRad2 := lon2*math.Pi/180, lat2*math.Pi/180

	deltaLon, deltaLat := lonRad2-lonRad1, latRad2-latRad1

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) + math.Cos(latRad1)*math.Cos(latRad2)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return EARTH_RADIUS_METERS * c
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

func (s *Store) GetDistanceMeters(key string, name1, name2 string) (float64, error) {
	lon1, lat1, err1 := s.GetLonLat(key, name1)
	lon2, lat2, err2 := s.GetLonLat(key, name2)
	if err1 != nil || err2 != nil {
		return 0, Error.ERR_MEMBER_NOT_FOUND
	}

	distance := HaversineDistance(lon1, lat1, lon2, lat2)
	return distance, nil
}

func (s *Store) GeoSearch(key string, lon, lat, radius float64) []string {
	redisObj, exists := s.data[key]
	if !exists {
		return []string{}
	}

	zsetObj := redisObj.Value.(*storage.SortedSet)

	var result []string = []string{}
	for name, score := range zsetObj.ScoreMap {
		lon2, lat2 := DecodeScore(int64(score))
		distance := HaversineDistance(lon, lat, lon2, lat2)
		if distance <= radius {
			result = append(result, name)
		}
	}
	return result
}
