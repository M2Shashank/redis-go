package persistence

import (
	"fmt"
	"os"
)

func CreateAppendOnlyDir(dir string) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Printf("Error creating appendonly directory: %v\n", err)
		return
	}
}