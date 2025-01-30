package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/f1-surya/git-go/initrepo"
	"github.com/f1-surya/git-go/staging"
)

func TestInit(t *testing.T) {
	os.RemoveAll(".git-go")

	initrepo.InitRepo()

	dirs := []string{
		".git-go",
		".git-go/refs",
		".git-go/refs/heads",
		".git-go/objects",
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
	initrepo.InitRepo()

	err := staging.Add([]string{"main.go", "main_test.go"})

	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	entries, err := staging.ReadIndex()
	if err != nil {
		t.Fatalf("Read index errored: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("Entries' missing files")
	}

	err = staging.Add([]string{"go.mod"})
	if err != nil {
		t.Fatalf("Adding to existing index errored, error: %v", err)
	}

	entries, err = staging.ReadIndex()
	if err != nil {
		t.Fatalf("Read index errored: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("Entries' missing 3rd file")
	}

	os.RemoveAll(".git-go")
	fmt.Println("TestAdd Passed")
}
