package main

import (
	"fmt"
	"os"

	"github.com/f1-surya/git-go/commands"
)

func checkRepo() error {
	if _, err := os.Stat(".git-go"); os.IsNotExist(err) {
		return fmt.Errorf("no repo initialized in this directory")
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a command")
		return
	}

	switch os.Args[1] {
	case "init":
		commands.Init()
	case "add":
		if err := checkRepo(); err == nil {
			if err := commands.Add(os.Args[2:]); err != nil {
				fmt.Println(fmt.Errorf("%w", err))
			}

		} else {
			fmt.Println("No repo initialized in this directory")
		}
	case "commit":
		if err := checkRepo(); err == nil {
			if err := commands.Commit(os.Args[2:]); err != nil {
				fmt.Println(fmt.Errorf("%w", err))
			}
		} else {
			fmt.Println("No repo initialized in this directory")
		}
	default:
		fmt.Println("Unknown command")
	}
}
