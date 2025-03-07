package commit

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"time"

	"github.com/f1-surya/git-go/object"
	"github.com/f1-surya/git-go/tree"
)

type Commit struct {
	Author    string
	CreatedAt time.Time
	Message   string
	Tree      string
	Parent    string
	Hash      string
}

func (c *Commit) ToBytes() ([]byte, error) {
	var buff bytes.Buffer
	buff.Write(fmt.Appendf(nil, "parent %s\n", c.Parent))
	buff.Write(fmt.Appendf(nil, "tree %s\n", c.Tree))

	user, err := user.Current()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	buff.Write(fmt.Appendf(nil, "author %v %d\n", user.Username, now.Unix()))
	buff.Write([]byte(c.Message))

	content := buff.Bytes()
	commit := fmt.Appendf(nil, "commit %d\n", len(content))
	commit = append(commit, content...)

	return commit, nil
}

func CreateCommit(args []string) (Commit, error) {
	var newCommit Commit
	root, err := tree.WriteTrees()
	if err != nil {
		return newCommit, err
	}

	newCommit.Tree = root
	newCommit.Message = args[1]

	headPath := filepath.Join(".git-go", "refs", "heads", "main")
	if _, err := os.Stat(headPath); err == nil {
		head, err := os.ReadFile(headPath)
		if err != nil {
			return newCommit, err
		}
		newCommit.Parent = string(head)
	}
	commitContent, err := newCommit.ToBytes()
	if err != nil {
		return newCommit, err
	}
	commitHash := sha1.Sum(commitContent)
	newCommit.Hash = hex.EncodeToString(commitHash[:])
	return newCommit, nil
}

// Writes the commit to the ObjectDB
func WriteCommit(commit Commit) error {
	commitBytes, err := commit.ToBytes()
	if err != nil {
		return err
	}
	err = object.WriteObject(commitBytes, commit.Hash)
	if err != nil {
		return err
	}

	tempHeadPath := filepath.Join(".git-go", "refs", "heads", "main.temp")
	tempHead, err := os.Create(tempHeadPath)
	if err != nil {
		return err
	}
	defer tempHead.Close()

	if _, err := tempHead.Write([]byte(commit.Hash)); err != nil {
		return err
	}
	if err = os.Rename(tempHeadPath, filepath.Join(".git-go", "refs", "heads", "main")); err != nil {
		return err
	}
	return nil
}

func ParseCommit(commitHash string) (Commit, error) {
	var commit Commit
	commitObject, err := object.ReadObject(commitHash)
	if err != nil {
		return commit, err
	}

	commitParts := bytes.Split(commitObject, []byte("\n"))

	spaceByte := []byte(" ")
	parent := bytes.Split(commitParts[1], spaceByte)[1]
	tree := bytes.Split(commitParts[2], spaceByte)[1]
	metadata := bytes.Split(commitParts[3], spaceByte)
	timestamp, err := strconv.ParseInt(string(metadata[2]), 10, 64)
	if err != nil {
		return commit, err
	}

	commit.Author = string(metadata[1])
	commit.CreatedAt = time.Unix(timestamp, 0)
	commit.Tree = string(tree)
	commit.Message = string(commitParts[4])
	commit.Parent = string(parent)
	commit.Hash = commitHash

	return commit, nil
}

func GetLatest() (*Commit, error) {
	var result Commit

	head, err := os.ReadFile(filepath.Join(".git-go", "refs", "heads", "main"))
	if err != nil {
		return nil, err
	}
	if len(head) == 0 {
		return nil, nil
	}
	result, err = ParseCommit(string(head))
	if err != nil {
		return nil, err
	}

	return &result, nil
}
