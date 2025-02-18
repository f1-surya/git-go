package commit_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/f1-surya/git-go/commands"
	"github.com/f1-surya/git-go/commit"
)

func TestParseCommit(t *testing.T) {
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

	err = commands.Add([]string{"commit.go", "test/test.txt"})
	if err != nil {
		t.Fatalf("Add errored: %v", err)
	}

	err = commands.Commit([]string{"-m", "Init"})
	if err != nil {
		t.Fatalf("Commit errored: %v", err)
	}

	head, err := os.ReadFile(filepath.Join(".git-go", "refs", "heads", "main"))
	if err != nil {
		t.Fatalf("Error while reading the HEAD: %v", err)
	}

	latestCommit, err := commit.ParseCommit(string(head))
	if err != nil {
		t.Fatalf("Error while parsing the commit: %v", err)
	}

	if latestCommit.Message != "Init" {
		t.Fatalf("Wrong message %s", latestCommit.Message)
	}

	file, err = os.Create("test/test-2.txt")
	if err != nil {
		t.Fatalf("File 2 creation errored: %v", err)
	}
	defer file.Close()

	err = commands.Add([]string{"test/test-2.txt"})
	if err != nil {
		t.Fatalf("Add 2 errored: %v", err)
	}

	err = commands.Commit([]string{"-m", "Second file"})
	if err != nil {
		t.Fatalf("Second commit errored: %v", err)
	}

	head, err = os.ReadFile(filepath.Join(".git-go", "refs", "heads", "main"))
	if err != nil {
		t.Fatalf("Error while reading the HEAD: %v", err)
	}

	latestCommit, err = commit.ParseCommit(string(head))
	if err != nil {
		t.Fatalf("Error while parsing the commit: %v", err)
	}

	if latestCommit.Message != "Second file" {
		t.Fatalf("Wrong message %s", latestCommit.Message)
	}

	if latestCommit.Parent == "" {
		t.Fatalf("Parent commit missing")
	}
}
