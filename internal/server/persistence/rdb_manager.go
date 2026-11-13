package persistence

import (
	"path/filepath"
	"os"
	"fmt"
)

type RDBManager struct {
	Data []byte
}

func NewRDBManager(dir, dbfilename string) *RDBManager {
	data := ReadRDBFile(dir, dbfilename)

	return &RDBManager{
		Data:      data,
	}
}

func ReadRDBFile(dir, dbfilename string) []byte {
	addr := filepath.Join(dir, dbfilename)

	data, err := os.ReadFile(addr)
	if err != nil {
		fmt.Println("Error reading RDB file: ", err.Error())
		return nil
	}
	fmt.Println("Read ", len(data), " bytes from RDB file: ", addr)
	return data
}