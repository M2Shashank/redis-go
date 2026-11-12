package rdb_parser

import (
	"encoding/binary"
	"fmt"
)

const AUX byte = 0xFA
const RESIZEDB byte = 0xFB
const EXPIRETIMEMS byte = 0xFC
const EXPIRETIME byte = 0xFD
const SELECTDB byte = 0xFE
const EOF byte = 0xFF

type ValueType struct {
	// Only handling string type for now.
	// For String , type = 0
	Type       byte
	ExpireTimeMS uint64
	Value      []byte
}

// value is string for now. Might have to change it to a struct later
func ParseRDBFile(data []byte) map[string]*ValueType {
	var headerSection []byte = data[0:9]
	fmt.Println("RDB Header: ", string(headerSection))
	data = data[9:] // remove header section

	_ = ReadAuxiliaryFields(data)

	mappings := ReadDataBaseSection(data)

	_ = ReadFooterSection(data)

	return mappings
}

func ReadAuxiliaryFields(data []byte) map[string]string {
	if len(data) < 1 || data[0] != AUX {
		fmt.Println("No Auxiliary Fields/Metadata found: expected 0xFA, got ", data[0])
		return nil
	}
	data = data[1:] // remove 0xFA

	AuxFields := make(map[string]string)
	for len(data) > 0 && (data[0] != EOF || data[0] != SELECTDB) {
		// Process each key-value pair in the auxiliary fields section
		key := ExtractString(data)
		value := ExtractString(data)

		AuxFields[key] = value
		fmt.Println("Auxiliary Field: ", key, " = ", value)
	}

	return AuxFields
}

func ReadDataBaseSection(data []byte) map[string]*ValueType {
	if len(data) < 2 || data[0] != SELECTDB {
		fmt.Println("No database section found: expected 0xFE, got ", data[0])
		return nil
	}
	_ = data[1]     // database number / index of database . 0 Here
	data = data[2:] // remove 0xFE

	if len(data) < 3 || data[0] != RESIZEDB {
		fmt.Println("No RESIZEDB section found: expected 0xFB, got ", data[0])
		return nil
	}
	noOfKeys, _ := int(data[1]), int(data[2])
	data = data[3:] // remove 0xFB and the two length bytes

	result := make(map[string]*ValueType, 0)
	for i := 0; i < noOfKeys; i++ {
		if len(data) > 1 {
			key, value := ExtractKeyValuePair(data)
			result[key] = value
		}
	}

	return result
}

func ExtractKeyValuePair(data []byte) (string, *ValueType) {
	var value ValueType = ValueType{
		ExpireTimeMS: 0,
		Type:       0,
		Value:      nil,
	}
	if data[0] == EXPIRETIMEMS {
		value.ExpireTimeMS = binary.BigEndian.Uint64(data[1:9])
		data = data[9:]
	} else if data[0] == EXPIRETIME {
		value.ExpireTimeMS = uint64(1000 * binary.BigEndian.Uint32(data[1:5]))
		data = data[5:]
	}
	value.Type = data[0]
	data = data[1:] // remove type byte

	key := ExtractString(data)
	value.Value = []byte(ExtractString(data))
	return key, &value
}

func ReadFooterSection(data []byte) int {
	if len(data) < 1 || data[0] != EOF {
		fmt.Println("No EOF section found: expected 0xFF, got ", data[0])
		return -1
	}

	checkSum := int(binary.LittleEndian.Uint64(data[1:9]))
	fmt.Println("RDB Footer: Checksum = ", checkSum)
	data = data[9:] // remove 0xFF and checksum bytes

	return checkSum
}






func ExtractLength(data []byte) (int, bool) {
	// Length Encoding:
	// First 2 bits : 00, 01, 10, 11
	// 00 -> (current byte) 6 bit length (0-63) in next 6 bits
	// 01 -> (current + next byte) 14 bit length (0-16383) in next 14 bits
	// 10 -> (discard current byte, read next 4 bytes) 32 bit length (0-4294967295) in next 32 bits
	// 11 -> special encoding. Remaining 6 bits indicate format.

	bits := data[0] & 0b11000000

	var length int = 0
	var special bool = false
	switch bits {
	case 0b00000000: // 6 bit length
		length = int(data[0] & 0b00111111)
		data = data[1:] // remove length byte
	case 0b01000000: // 14 bit length
		length = int((data[0]&0b00111111)<<8) | int(data[1])
		data = data[2:] // remove length bytes
	case 0b10000000: // 32 bit length
		length = int(binary.BigEndian.Uint32(data[1:5]))
		data = data[5:] // remove length bytes
	case 0b11000000: // special encoding.
		length = int(data[0] & 0b00111111)
		special = true
		data = data[1:] // remove length byte
	}
	return length, special
}

func ExtractString(data []byte) string {
	// String Encoding: [len][string bytes]

	// For Interger in String. The first 2 bits of length == 11.
	// Remaining 6 bits indicate the format of the integer
	// 0 -> 8 bit signed integer
	// 1 -> 16 bit signed integer
	// 2 -> 32 bit signed integer
	// 3 -> LZF compressed string. Not handled for now.

	if len(data) < 1 {
		fmt.Println("Error: data is empty")
		return ""
	}

	len, special := ExtractLength(data)
	var str string
	if !special {
		str = string(data[0:len])
		data = data[len:] // remove string bytes
	} else {
		// Special Encoding, String is a Integer.
		var integer int
		if len == 0 {
			integer = int(int8(data[0]))
			data = data[1:] // remove integer byte
		} else if len == 1 {
			integer = int(int16(binary.LittleEndian.Uint16(data[0:2])))
			data = data[2:] // remove integer bytes
		} else if len == 2 {
			integer = int(int32(binary.LittleEndian.Uint32(data[0:4])))
			data = data[4:] // remove integer bytes
		} else {
			// len == 3
			// LZF compressed string. Not handled for now.
		}
		str = fmt.Sprintf("%d", integer)
	}

	return str
}
