package storage

type String struct {
	Data   []byte
	Length int // length of the string in bytes
}


// String Store Handler Functions

func (s *String) SetData(value string) {
	s.Data = []byte(value)
	s.Length = len(value)
}

// Bitmap Handler Functions on String

func (s *String) SetBitVal(byteIndex, bitIndex, value int) int {
	if byteIndex+1 > s.Length {
		s.Data = append(
			s.Data,
			make([]byte, byteIndex-s.Length+1)...,
		)
		s.Length = byteIndex + 1
	}

	dataByte := &s.Data[byteIndex]
	prevBit := (*dataByte >> (7 - bitIndex)) & 1

	mask := byte(1 << (7 - bitIndex))
	if value == 1 {
		*dataByte |= mask
	} else {
		*dataByte &= (^mask)
	}

	return int(prevBit)
}

func (s *String) GetBitVal(byteIndex, bitIndex int) int {
	if byteIndex+1 > s.Length {
		return 0
	}

	return int((s.Data[byteIndex] >> (7 - bitIndex)) & 1)
}

func (s *String) CountSetBits(start, end int) int {
	end = min(end, s.Length-1)
	if end == -1 {
		end = s.Length - 1
	}
	if start < 0 || end >= s.Length || start > end {
		return 0
	}

	var count int = 0
	for i := start; i <= end; i++ {
		for j := range 8 {
			count += int((s.Data[i] >> j) & 1)
		}
	}
	return count
}

func (s *String) AndOne(o *String) {
	minLength := min(s.Length, o.Length)
	for i := range minLength {
		s.Data[i] &= o.Data[i]
	}

	if o.Length > s.Length {
		s.Data = append(s.Data, make([]byte, o.Length-s.Length)...)
		s.Length = o.Length
	} else {
		clear(s.Data[minLength:s.Length])
	}
}

func (s *String) AndAll(keys []*String) int {
	if len(keys) == 0 {
		s.Data = []byte{0}
		s.Length = 1
		return 1
	}

	s.Data = append([]byte{}, keys[0].Data...)
	s.Length = keys[0].Length

	for _, key := range keys {
		s.AndOne(key)
	}
	return s.Length
}

func (s *String) OrOne(o *String) {
	minLength := min(s.Length, o.Length)
	for i := range minLength {
		s.Data[i] |= o.Data[i]
	}

	if o.Length > s.Length {
		s.Data = append(s.Data, o.Data[s.Length:]...)
		s.Length = o.Length
	}
}

func (s *String) OrAll(keys []*String) int {
	if len(keys) == 0 {
		s.Data = []byte{0}
		s.Length = 1
		return 1
	}

	s.Data = append([]byte{}, keys[0].Data...)
	s.Length = keys[0].Length

	for _, key := range keys {
		s.OrOne(key)
	}
	return s.Length
}