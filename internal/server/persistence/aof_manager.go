package persistence

import (
	"fmt"
	"os"
	"path/filepath"
)

type AOFManager struct {
	AofFile  *os.File
	Manifest *os.File
	AOFAddr  string
}

func NewAOFManager(dir string, appendfilename string) *AOFManager {
	aofAddr := filepath.Join(dir, appendfilename)

	aofFileName := appendfilename + ".1.incr.aof"
	file := CreateAppendOnlyFile(dir, aofFileName)

	manFileName := appendfilename + ".manifest"
	manifest := CreateManifestFile(filepath.Join(dir, manFileName), aofFileName)

	return &AOFManager{
		AofFile:  file,
		AOFAddr:  aofAddr,
		Manifest: manifest,
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

func CreateManifestFile(filePath string, aofFileName string) *os.File {
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Error creating manifest file: %v\n", err)
		return nil
	}

	temp := fmt.Sprintf("file %s seq 1 type i", aofFileName)
	data := []byte(temp)

	_, err = file.Write(data)
	if err != nil {
		fmt.Printf("Error writing to manifest file: %v\n", err)
		return nil
	}

	return file
}
