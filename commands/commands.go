package commands

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/f1-surya/git-go/index"
)

func Init() {
	dirs := []string{
		filepath.Join(".git-go", "refs", "heads"),
		filepath.Join(".git-go", "objects"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			failed := fmt.Errorf("failed to create %s: %v", dir, err)
			fmt.Println(failed)
			return
		}
	}

	indexFile, err := os.Create(filepath.Join(".git-go", "index"))
	if err != nil {
		fmt.Printf("Error while creating the index file error: %v", err)
		return
	}
	defer indexFile.Close()

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

func Add(files []string) error {
	if len(files) == 0 {
		return errors.New("no files are provided to stage")
	}

	uniqueEntries := make(map[string]index.IndexEntry)

	oldEntries, err := index.ReadIndex()
	if err != nil {
		return err
	}

	for _, oldEntry := range oldEntries {
		uniqueEntries[oldEntry.Path] = oldEntry
	}

	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("file does not exist %s", file)
		}

		fileContent, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		uniqueEntries[file] = index.IndexEntry{
			Mode:    0o100644,
			Size:    uint32(len(fileContent)),
			Hash:    sha1.Sum(fileContent),
			Path:    file,
			Content: fileContent,
		}
	}

	var entries []index.IndexEntry
	for _, entry := range uniqueEntries {
		entries = append(entries, entry)
	}

	sort.Sort(index.ByPath(entries))
	return index.WriteIndex(entries)
}
