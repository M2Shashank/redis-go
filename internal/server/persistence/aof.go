package persistence

import (
	"fmt"
	"os"
	"path/filepath"
)

type AOFHandler struct {
	File *os.File
	DirName, FileName string
	AOFAddr string
}

func NewAOFHandler(dirName string, fileName string) *AOFHandler {
	aofAddr := filepath.Join(dirName, fileName)
	file := CreateAppendOnlyFile(dirName, fileName)

	return &AOFHandler{
		File: file,
		DirName: dirName,
		FileName: fileName,
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
