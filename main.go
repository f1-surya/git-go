package main

import (
	"fmt"
	"os"
)

func initRepo() {
	dirs := []string{
		".git-go",
		".git-go/refs",
		".git-go/refs/heads",
		".git-go/objects",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			failed := fmt.Errorf("failed to create %s: %v", dir, err)
			fmt.Println(failed)
			return
		}
	}

	fmt.Println("Successfully created repo")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a command")
		return
	}

	switch os.Args[1] {
	case "init":
		initRepo()
	default:
		fmt.Println("Unknown command")
	}
}
