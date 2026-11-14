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
	Manifest *os.File
	CurAofFile *os.File
}

// Creates a new AOFManager and return it. 
// If a manifest file already exists, it opens the existing manifest file and retrieves the current AOF file name from it.
// If manifest exists, it also returns the file descriptor of the current AOF file. 
// If manifest doesn't exist, it returns null
func NewAOFManager(dir string, appendfilename string) (*AOFManager, []byte) {
	// Do not close. Sent Open file descriptor of manifest
	manFileName := appendfilename + ".manifest"
	manifest, exists := CreateManifestFile(dir, manFileName)
	if manifest == nil {
		fmt.Println("Error creating manifest file. AOF will not be created.")
		return nil, nil
	}

	if !exists {
		// Create the AOF and add it to the manifest file
		aofFileName := appendfilename + ".1.incr.aof"
		err := CreateAppendOnlyFile(dir, aofFileName, manifest)
		if err != nil {
			fmt.Printf("Error creating AOF file: %v\n", err)
			return nil, nil
		}
	}

	curAOFfilename := GetFileNameFromManifest(manifest)
	curaofFilePath := filepath.Join(dir, curAOFfilename)
	curAofFile, err := os.OpenFile(curaofFilePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening current AOF file: %v\n", err)
		return nil, nil
	}

	fileData, err := os.ReadFile(curaofFilePath)
	if err != nil {
		fmt.Printf("Error reading current AOF file: %v\n", err)
		return nil, nil
	}

	manager := &AOFManager{
		Directory:      dir,
		Appendfilename: appendfilename,
		Manifest:       manifest,
		CurAofFile:     curAofFile,
	}

	if exists {
		return manager, fileData
	}
	return manager, nil
}

// Opens manifest file. If it doesn't exist, creates a new one. 
// Returns the file descriptor of the manifest file, if it already existed.
func CreateManifestFile(dir string, filename string) (*os.File, bool) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Printf("Error creating appendonly directory: %v\n", err)
		return nil, false
	}

	filePath := filepath.Join(dir, filename)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)
	if err == nil {
		return file, false // New manifest file created
	}

	if os.IsExist(err) {
		file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
		if err != nil {
			fmt.Printf("Error opening existing manifest file: %v\n", err)
			return nil, false // Manifest file already exists
		}
		return file, true // Manifest file already exists
	}
	return nil, false
}

// Creates a new append-only file and adds it to the manifest.
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

// A file's addr can be retrieved using [os.File].name()
func (aof *AOFManager) WriteCommand(cmd []byte, sync string) {
	// Get which file to write in from the manifest file
	// aofFileName := GetFileNameFromManifest(aof.Manifest)

	// Read current AOF file name from the manifest file during server startup. 
	// Using the stored current aof
	if aof.CurAofFile == nil {
		fmt.Println("Error: Could not retrieve AOF file name from manifest.")
		return
	}

	// Write the command to the AOF file
	_, err := aof.CurAofFile.Write(cmd)
	if err != nil {
		fmt.Printf("Error writing to AOF file: %v\n", err)
		return
	}

	if sync == "always" {
		if err := aof.CurAofFile.Sync(); err != nil {
			fmt.Printf("Error syncing AOF: %v\n", err)
		}
	}
}

// Instead of reading the manifest for every write command, we can read it once when the AOFManager is created and store the current AOF file name in a field.
func GetFileNameFromManifest(manifest *os.File) string {
	if manifest == nil {
		return ""
	}

	data, err := os.ReadFile(manifest.Name())
	if err != nil {
		fmt.Println("Error reading manifest file:", err)
		return ""
	}

	var filename string = ""
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) == 6 && parts[4] == "type" && parts[5] == "i" {
			filename = parts[1] // Get latest AOF file name
		}
	}
	return filename
}
