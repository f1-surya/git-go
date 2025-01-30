package staging

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/f1-surya/git-go/utils"
)

type IndexEntry struct {
	Mode uint32
	Size uint32
	Hash [20]byte
	Path string
}

type ByPath []IndexEntry

func (a ByPath) Len() int           { return len(a) }
func (a ByPath) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPath) Less(i, j int) bool { return a[i].Path < a[j].Path }

func ReadIndex() ([]IndexEntry, error) {
	indexFile, err := os.Open(".git-go/index")
	if err != nil {
		return nil, err
	}
	defer indexFile.Close()

	header := make([]byte, 4)
	if _, err := indexFile.Read(header); err != nil {
		return nil, fmt.Errorf("error while parsing header: %v", err)
	}
	if string(header) != "DIRC" {
		return nil, fmt.Errorf("invalid index format")
	}

	var entryCount uint32
	if err := binary.Read(indexFile, binary.BigEndian, &entryCount); err != nil {
		return nil, fmt.Errorf("error while parsing entries count: %v", err)
	}

	var entries []IndexEntry

	for i := uint32(0); i < entryCount; i++ {
		var entry IndexEntry

		if err := binary.Read(indexFile, binary.BigEndian, &entry.Mode); err != nil {
			return nil, fmt.Errorf("could not parse entry mode: %v", err)
		}

		if err := binary.Read(indexFile, binary.BigEndian, &entry.Size); err != nil {
			return nil, fmt.Errorf("could not parse entry size: %v", err)
		}

		if _, err := indexFile.Read(entry.Hash[:]); err != nil {
			return nil, fmt.Errorf("could not parse entry hash: %v", err)
		}

		path := ""
		for {
			b := make([]byte, 1)
			if _, err := indexFile.Read(b); err != nil {
				return nil, fmt.Errorf("could not parse entry path: %v", err)
			}
			if b[0] == 0 {
				break
			}
			path += string(b)
		}
		entry.Path = path
		entries = append(entries, entry)
	}

	return entries, nil
}

func WriteIndex(entries []IndexEntry) error {
	indexFile, err := os.Create(".git-go/index.temp")
	if err != nil {
		return err
	}
	defer indexFile.Close()

	header := []byte("DIRC")
	if _, err := indexFile.Write(header); err != nil {
		fmt.Println("could not write the header: %w", err)
		return err
	}

	if err := binary.Write(indexFile, binary.BigEndian, uint32(len(entries))); err != nil {
		return fmt.Errorf("error while writing the length: %w", err)
	}

	for _, entry := range entries {
		if err := binary.Write(indexFile, binary.BigEndian, entry.Mode); err != nil {
			return fmt.Errorf("error while writing entry mode: %w", err)
		}
		if err := binary.Write(indexFile, binary.BigEndian, entry.Size); err != nil {
			return fmt.Errorf("error while writing entry size: %w", err)
		}
		if _, err := indexFile.Write(entry.Hash[:]); err != nil {
			return fmt.Errorf("error while writing entry hash: %w", err)
		}
		if _, err := indexFile.WriteString(entry.Path + "\x00"); err != nil {
			return fmt.Errorf("error while writing entry path: %w", err)
		}
	}

	if err := os.Rename(".git-go/index.temp", ".git-go/index"); err != nil {
		return err
	}

	return nil
}

func Add(files []string) error {
	if len(files) == 0 {
		return errors.New("no files are provided to stage")
	}

	var entries []IndexEntry
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("file does not exist %s", file)
		}

		hash, size, err := utils.HashFile(file)
		if err != nil {
			return err
		}

		entry := IndexEntry{
			Mode: 0o100644,
			Size: size,
			Hash: hash,
			Path: file,
		}
		entries = append(entries, entry)
	}

	oldEntries, err := ReadIndex()
	if err != nil {
		return err
	}

	entries = append(entries, oldEntries...)
	sort.Sort(ByPath(entries))
	return WriteIndex(entries)
}
