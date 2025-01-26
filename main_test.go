package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	os.RemoveAll(".git-go")

	initRepo()

	dirs := []string{
		".git-go",
		".git-go/refs",
		".git-go/refs/heads",
		".git-go/objects",
	}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf(dir + " not found")
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
