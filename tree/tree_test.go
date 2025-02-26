package tree_test

import (
	"os"
	"testing"

	"github.com/f1-surya/git-go/commands"
	"github.com/f1-surya/git-go/commit"
	"github.com/f1-surya/git-go/tree"
)

func TestGetTreesRecursive(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll(".git-go")
		os.RemoveAll("test")
	})
	commands.Init()

	err := os.Mkdir("test", 0755)
	if err != nil {
		t.Fatalf("Dir creation errored: %v", err)
	}

	file, err := os.Create("test/test.txt")
	if err != nil {
		t.Fatalf("File creation errored: %v", err)
	}
	defer file.Close()

	err = commands.Add([]string{"tree.go", "test/test.txt"})
	if err != nil {
		t.Fatalf("Add errored: %v", err)
	}

	err = commands.Commit([]string{"-m", "Init"})
	if err != nil {
		t.Fatalf("Commit errored: %v", err)
	}
	latestCommit, err := commit.GetLatest()
	if err != nil {
		t.Fatalf("Failed to fetch latest commit: %v", err)
	}

	trees, err := tree.GetTreesRecursive(latestCommit.Tree)
	if err != nil {
		t.Fatalf("Failed to pares root: %v", err)
	}
	if len(trees["."].Children) != 2 {
		t.Fatalf("Incorrect number of children in tree root: %d", len(trees["."].Children))
	}
	if len(trees["test"].Children) != 1 {
		t.Fatalf("Incorrect number of children in test tree: %d", len(trees["test"].Children))
	}
}
