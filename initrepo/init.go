package initrepo

import (
	"encoding/binary"
	"fmt"
	"os"
)

func InitRepo() {
	dirs := []string{
		".git-go",
		".git-go/refs",
		".git-go/refs/heads",
		".git-go/objects",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			failed := fmt.Errorf("failed to create %s: %v", dir, err)
			fmt.Println(failed)
			return
		}
	}

	indexFile, err := os.Create(".git-go/index")
	if err != nil {
		fmt.Printf("Error while creating the index file error: %v", err)
		return
	}

	header := []byte("DIRC")
	if _, err := indexFile.Write(header); err != nil {
		fmt.Printf("Error while writing index header, error: %v", err)
		return
	}

	if err := binary.Write(indexFile, binary.BigEndian, uint32(0)); err != nil {
		fmt.Printf("Error while writing index entry count, error: %v", err)
		return
	}

	fmt.Println("Successfully created repo")
}
