package store

func CalcScore(lat, lon float64) float64 {
	return 0
}

func (s *Store) GEOADD(key string, lat, lon float64, name string) (bool, error) {
	score := CalcScore(lat, lon)
	return s.ZADD(key, score, name)
}