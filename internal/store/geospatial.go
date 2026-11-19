package store

import (
	"fmt"
)

func (s *Store) GEOADD(key string, lon, lat float64, name string) (bool, error) {
	score := EncodeScore(lon, lat)
	fmt.Printf("Encoded Score for (%f, %f) is %d\n", lon, lat, score)
	return s.ZADD(key, float64(score), name)
}

func EncodeScore(lon, lat float64) int64 {
	// Normalisation to range of [0, 2^26) .
	normalisedLon := (1 << 26) * (lon + 180) / 360
	normalisedLat := (1 << 26) * (lat + 85.05112878) / 170.10225756

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
	var lon float64 = float64(intLon)*360/(2<<26) - 180
	var lat float64 = float64(intLat)*170.10225756/(2<<26) - 85.05112878

	return lon, lat
}
