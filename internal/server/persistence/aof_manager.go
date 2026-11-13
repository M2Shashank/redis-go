package persistence

import (
	"fmt"
	"os"
	"path/filepath"
)

type AOFManager struct {
	File *os.File
	AOFAddr string
}

func NewAOFManager(directory string, fileName string) *AOFManager {
	aofAddr := filepath.Join(directory, fileName)
	file := CreateAppendOnlyFile(directory, fileName)

	return &AOFManager{
		File: file,
		AOFAddr: aofAddr,
	}
}

func CreateAppendOnlyFile(dir string, filename string) *os.File {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Printf("Error creating appendonly directory: %v\n", err)
		return nil
	}

	filePath := filepath.Join(dir, filename)
	file, err1 := os.Create(filePath)
	if err1 != nil {
		fmt.Printf("Error creating appendonly file: %v\n", err)
		return nil
	}

	return file
}
