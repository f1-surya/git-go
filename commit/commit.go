package commit

import (
	"bytes"
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

func CreateCommit(args []string) ([]byte, error) {
	root, err := tree.WriteTrees()
	if err != nil {
		return nil, err
	}

	var commit []byte
	headPath := filepath.Join(".git-go", "refs", "heads", "main")
	if _, err := os.Stat(headPath); err == nil {
		head, err := os.ReadFile(headPath)
		if err != nil {
			return nil, err
		}
		commit = append(commit, fmt.Appendf(nil, "parent %s\n", string(head))...)
	}
	commit = append(commit, fmt.Appendf(nil, "tree %s\n", root)...)

	user, err := user.Current()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	commit = append(commit, fmt.Appendf(nil, "author %v %d\n", user.Username, now.Unix())...)
	commit = append(commit, []byte(args[1])...)
	commit = append(fmt.Appendf(nil, "commit %d\n", len(commit)), commit...)

	return commit, nil
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
