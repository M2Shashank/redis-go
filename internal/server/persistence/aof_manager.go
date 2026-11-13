package persistence

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Path :  dir/appendfilename.{ manifest, 1.incr.aof }
type AOFManager struct {
	Directory, Appendfilename string
	Manifest                  *os.File
}

func NewAOFManager(dir string, appendfilename string) *AOFManager {
	// Do not close. Sent Open file descriptor of manifest
	manFileName := appendfilename + ".manifest"
	manifest := CreateManifestFile(dir, manFileName)

	// Create the AOF and add it to the manifest file
	aofFileName := appendfilename + ".1.incr.aof"
	err := CreateAppendOnlyFile(dir, aofFileName, manifest)
	if err != nil {
		fmt.Printf("Error creating AOF file: %v\n", err)
		return nil
	}

	return &AOFManager{
		Directory:      dir,
		Appendfilename: appendfilename,
		Manifest:       manifest,
	}
}

func CreateManifestFile(dir string, filename string) *os.File {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Printf("Error creating appendonly directory: %v\n", err)
		return nil
	}

	filePath := filepath.Join(dir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Error creating manifest file: %v\n", err)
		return nil
	}
	return file
}

func CreateAppendOnlyFile(dir string, filename string, manifest *os.File) error {
	filePath := filepath.Join(dir, filename)
	aofFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("Failed to create or open AOF file: %v\n", err)
	}
	aofFile.Close() // Close the file after creating it

	data := fmt.Sprintf("file %s seq 1 type i\n", filename)
	_, err = manifest.Write([]byte(data))
	if err != nil {
		return fmt.Errorf("Error writing to manifest file: %v\n", err)
	}

	return nil
}

// I dont need the AOFFile and AOFAddr in the AOFManager struct.
// I just need the directory and the manifest file.
// The current AOF file should be retrieved from the manifest file.
// A files addr can be retrieved using [os.File].name()
func (aof *AOFManager) WriteCommand(cmd []byte, sync string) {
	// Get which file to write in from the manifest file
	aofFileName := aof.GetFileNameFromManifest()
	if aofFileName == "" {
		fmt.Println("Error: Could not retrieve AOF file name from manifest.")
		return
	}

	// Write the command to the AOF file
	aofFilePath := filepath.Join(aof.Directory, aofFileName)
	file, err := os.OpenFile(aofFilePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening AOF file: %v\n", err)
		return
	}
	defer file.Close()

	_, err = file.Write(cmd)
	if err != nil {
		fmt.Printf("Error writing to AOF file: %v\n", err)
		return
	}

	if sync == "always" {
		if err := file.Sync(); err != nil {
			fmt.Printf("Error syncing AOF: %v\n", err)
		}
	}
}

// Instead of reading the manifest for every write command, we can read it once when the AOFManager is created and store the current AOF file name in a field.

func (aof *AOFManager) GetFileNameFromManifest() string {
	if aof.Manifest == nil {
		return ""
	}

	data, err := os.ReadFile(aof.Manifest.Name())
	if err != nil {
		fmt.Println("Error reading manifest file:", err)
		return ""
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) == 6 && parts[4] == "type" && parts[5] == "i" {
			return parts[1] // Return the AOF file name
		}
	}
	return ""
}
