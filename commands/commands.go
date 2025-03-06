package commands

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/f1-surya/git-go/commit"
	"github.com/f1-surya/git-go/index"
	"github.com/f1-surya/git-go/object"
	"github.com/f1-surya/git-go/tree"
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

func Status() error {
	allFiles := make(map[string]string)
	var filesInCommit map[string]string
	indexEntries, err := index.ReadIndex()
	if err != nil {
		return err
	}
	entries := make(map[string]index.IndexEntry)

	for _, entry := range indexEntries {
		entries[entry.Path] = entry
		allFiles[entry.Path] = hex.EncodeToString(entry.Hash[:])
	}

	latestCommit, err := commit.GetLatest()
	if err != nil {
		return err
	}
	if latestCommit != nil {
		trees, err := tree.GetTreesRecursive(latestCommit.Tree)
		if err != nil {
			return fmt.Errorf("error while parsing root: %v", err)
		}
		filesInCommit = tree.GetAllFiles(trees)
		maps.Copy(allFiles, filesInCommit)
	}

	walkedFiles := make(map[string]string)

	var wg sync.WaitGroup
	var mu sync.Mutex
	err = filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if strings.Contains(path, ".git") {
				return nil
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				fileContent, err := os.ReadFile(path)
				if err != nil {
					fmt.Println("err")
					fmt.Printf("reading %s errored, e: %v", path, err)
					return
				}
				fileHash := sha1.Sum(fileContent)
				hashString := hex.EncodeToString(fileHash[:])
				mu.Lock()
				walkedFiles[path] = hashString
				allFiles[path] = hashString
				mu.Unlock()
			}()
		}
		return nil
	})
	wg.Wait()
	if err != nil {
		return err
	}

	var staged []string
	var notStaged []string

	for file := range allFiles {
		fsHash, inFs := walkedFiles[file]
		commitHash, inCommit := filesInCommit[file]
		indexEntry, inIndex := entries[file]
		indexHash := hex.EncodeToString(indexEntry.Hash[:])

		if fsHash == indexHash && indexHash == commitHash {
			continue
		}
		if !inFs && inIndex && inCommit {
			notStaged = append(notStaged, "deleted: "+file)
		} else if !inFs && !inIndex && inCommit {
			staged = append(staged, "deleted: "+file)
		} else if fsHash == indexHash && !inCommit {
			staged = append(staged, "created: "+file)
		} else if fsHash == indexHash && indexHash != commitHash {
			staged = append(staged, "modified: "+file)
		} else if fsHash != indexHash && !inIndex {
			notStaged = append(notStaged, "created: "+file)
		} else if fsHash != indexHash && inIndex {
			notStaged = append(notStaged, "modified: "+file)
		}
	}

	sort.Strings(staged)
	sort.Strings(notStaged)

	if len(staged) == 0 && len(notStaged) == 0 {
		fmt.Println("No changes detected")
		return nil
	}

	if len(staged) > 0 {
		fmt.Println("Changes staged for commit:\033[32m")
		fmt.Println("")
		for _, path := range staged {
			fmt.Println("    " + path)
		}
		fmt.Println("\033[0m")
	}

	if len(notStaged) > 0 {
		fmt.Println("Changes not staged for commit:\033[31m")
		fmt.Println("")
		for _, path := range notStaged {
			fmt.Println("    " + path)
		}
		fmt.Println("\033[0m")
	}

	return nil
}
