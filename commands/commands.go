package commands

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/f1-surya/git-go/commit"
	"github.com/f1-surya/git-go/index"
	"github.com/f1-surya/git-go/object"
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

	headFile, err := os.Create(filepath.Join(".git-go", "refs", "heads", "main"))
	if err != nil {
		fmt.Printf("Error while creating the head file: %v", err)
		return
	}
	defer headFile.Close()

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

// Adds the entered files to index and creates objects for them.
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
			_, ok := uniqueEntries[file]
			if ok {
				delete(uniqueEntries, file)
				continue
			}
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
			Path:    filepath.Clean(file),
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

func Commit(args []string) error {
	if len(args) < 2 {
		return errors.New("missing commit message")
	}

	commit, err := commit.CreateCommit(args)
	if err != nil {
		return err
	}
	commitHash := sha1.Sum(commit)
	commitHashString := hex.EncodeToString(commitHash[:])
	err = object.WriteObject(commit, commitHashString)
	if err != nil {
		return err
	}

	tempHeadPath := filepath.Join(".git-go", "refs", "heads", "main.temp")
	tempHead, err := os.Create(tempHeadPath)
	if err != nil {
		return err
	}
	defer tempHead.Close()

	if _, err := tempHead.Write([]byte(commitHashString)); err != nil {
		return err
	}
	if err = os.Rename(tempHeadPath, filepath.Join(".git-go", "refs", "heads", "main")); err != nil {
		return err
	}

	return nil
}

func Log() error {
	head := ""
	var commits []commit.Commit
	if headBytes, err := os.ReadFile(filepath.Join(".git-go", "refs", "heads", "main")); err != nil {
		return err
	} else {
		head = string(headBytes)
	}

	if head == "" {
		fmt.Println("There are no commits yet")
		return nil
	}

	for head != "" {
		currCommit, err := commit.ParseCommit(head)
		if err != nil {
			return err
		}
		commits = append([]commit.Commit{currCommit}, commits...)
		head = currCommit.Parent
	}

	for _, currCommit := range commits {
		fmt.Printf("\033[33mcommit %s\n\033[0m", currCommit.Hash)
		fmt.Printf("Author: %s\n", currCommit.Author)
		fmt.Printf("Date: %s\n\n", currCommit.CreatedAt.Format("Mon Jan 2 15:04:05 2006 MST"))
		fmt.Printf("   %s\n\n", currCommit.Message)
	}
	return nil
}
