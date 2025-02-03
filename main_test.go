package main_test

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/f1-surya/git-go/commands"
	"github.com/f1-surya/git-go/index"
)

func TestInit(t *testing.T) {
	os.RemoveAll(".git-go")

	commands.Init()

	dirs := []string{
		".git-go",
		filepath.Join(".git-go", "refs"),
		filepath.Join(".git-go", "refs", "heads"),
		filepath.Join(".git-go", "objects"),
	}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("%s not found", dir)
		}
	}

	os.RemoveAll(".git-go")
}

func TestInitCLI(t *testing.T) {
	os.RemoveAll(".git-go")

	command := exec.Command("go", "run", "main.go", "init")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	if !strings.Contains(string(output), "Successfully created repo") {
		t.Errorf("Got wrong output")
	}

	os.RemoveAll(".git-go")
}

func TestAdd(t *testing.T) {
	os.RemoveAll(".git-go")
	commands.Init()

	err := commands.Add([]string{"main.go", "main_test.go"})

	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	entries, err := index.ReadIndex()
	if err != nil {
		t.Fatalf("Read index errored: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("Entries' missing files")
	}

	err = commands.Add([]string{"go.mod"})
	if err != nil {
		t.Fatalf("Adding to existing index errored, error: %v", err)
	}

	entries, err = index.ReadIndex()
	if err != nil {
		t.Fatalf("Read index errored: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("Entries' missing 3rd file")
	}

	err = commands.Add([]string{"go.mod"})
	if err != nil {
		t.Fatalf("Adding to existing index the 2nd time errored, error: %v", err)
	}

	entries, err = index.ReadIndex()
	if err != nil {
		t.Fatalf("Read index errored: %v", err)
	}
	path := filepath.Join(".git-go", "objects", hex.EncodeToString(entries[0].Hash[:])[38:], hex.EncodeToString(entries[0].Hash[:]))
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Object doesn't exist")
	}

	if len(entries) != 3 && len(entries) == 4 {
		t.Fatalf("duplicate entry is present")
	}

	os.RemoveAll(".git-go")
	fmt.Println("TestAdd Passed")
}
