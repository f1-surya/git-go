package commands_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/f1-surya/git-go/commands"
	"github.com/f1-surya/git-go/commit"
	"github.com/f1-surya/git-go/index"
)

func TestCommitWithoutMessage(t *testing.T) {
	err := commands.Commit([]string{})
	if err == nil {
		t.Fatalf("No args are provided but the function didn't return any errors")
	}
}

func TestCommit(t *testing.T) {
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

	err = commands.Add([]string{"commands.go", "test/test.txt"})
	if err != nil {
		t.Fatalf("Add errored: %v", err)
	}

	err = commands.Commit([]string{"-m", "Init"})
	if err != nil {
		t.Fatalf("Commit errored: %v", err)
	}
	objectCount := 0
	err = filepath.Walk(filepath.Join(".git-go", "objects"), func(path string, info fs.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			objectCount++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Counting objects errored: %v", err)
	}
	if objectCount != 5 {
		t.Fatalf("Objects count is incorrect: %d", objectCount)
	}
}

func TestLog(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll(".git-go")
		os.RemoveAll("test")
	})

	commands.Init()
	err := commands.Add([]string{"commands.go"})
	if err != nil {
		t.Fatalf("Add errored: %v", err)
	}
	err = commands.Commit([]string{"-m", "Init"})
	if err != nil {
		t.Fatalf("Commit errored: %v", err)
	}

	err = os.Mkdir("test", 0755)
	if err != nil {
		t.Fatalf("Dir creation errored: %v", err)
	}
	file, err := os.Create("test/test.txt")
	if err != nil {
		t.Fatalf("File creation errored: %v", err)
	}
	defer file.Close()
	err = commands.Add([]string{"test/test.txt"})
	if err != nil {
		t.Fatalf("Add 2 errored: %v", err)
	}

	err = commands.Commit([]string{"-m", "Second"})
	if err != nil {
		t.Fatalf("Commit 2 errored: %v", err)
	}

	err = commands.Log()
	if err != nil {
		t.Fatalf("Log errored: %v", err)
	}
}

func TestDeletes(t *testing.T) {
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
	err = commands.Add([]string{"commands.go", "test/test.txt"})
	if err != nil {
		t.Fatalf("Add 2 errored: %v", err)
	}

	err = os.RemoveAll("test/test.txt")
	if err != nil {
		t.Fatalf("Deleting test errored: %v", err)
	}

	err = commands.Add([]string{"test/test.txt"})
	if err != nil {
		t.Fatalf("Add 2 errored: %v", err)
	}

	entries, err := index.ReadIndex()
	if err != nil {
		t.Fatalf("Read index errored: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("Entries length isn't right: %d", len(entries))
	}
}

func TestStatus(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll(".git-go")
		os.RemoveAll("test")
	})
	os.Chdir("..")
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

	err = commands.Add([]string{"commands/commands.go", "test/test.txt"})
	if err != nil {
		t.Fatalf("Add errored: %v", err)
	}

	err = commands.Status()
	if err != nil {
		t.Fatalf("status errored: %v", err)
	}
}

func TestRevert(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll(".git-go")
		os.RemoveAll("test")
	})

	os.Chdir("..")
	os.Mkdir("test", 0755)

	// Init
	commands.Init()

	// Create new files in test and commit them.
	err := os.WriteFile("test/hi.txt", []byte("hi"), 0644)
	if err != nil {
		t.Fatalf("file creation errored: %v", err)
	}
	err = commands.Add([]string{"test/hi.txt"})
	if err != nil {
		t.Fatalf("add errored: %v", err)
	}
	err = commands.Commit([]string{"-m", "Init"})
	if err != nil {
		t.Fatalf("commit errored: %v", err)
	}

	// Create another file in test, add some content to the previous file and commit them.
	err = os.WriteFile("test/hello.txt", []byte("hello"), 0644)
	if err != nil {
		t.Fatalf("hello creation errored: %v", err)
	}
	err = os.WriteFile("test/hi.txt", []byte("hello"), 0644)
	if err != nil {
		t.Fatalf("hi modification errored: %v", err)
	}
	err = commands.Add([]string{"test/hello.txt", "test/hi.txt"})
	if err != nil {
		t.Fatalf("add hello errored: %v", err)
	}
	err = commands.Commit([]string{"-m", "hello"})
	if err != nil {
		t.Fatalf("commit hello errored: %v", err)
	}

	// Call revert and check if the file that was added later exists or not.
	commands.Revert()
	// Read the content of the first file and make sure it matches the content before the second commit.
	hiContent, err := os.ReadFile(filepath.Join("test", "hi.txt"))
	if err != nil {
		t.Fatalf("error while reading hi.txt: %v", err)
	}
	if string(hiContent) == "hello" {
		t.Fatalf("Wrong content in hi.txt: %s", string(hiContent))
	}

	_, err = os.ReadFile("test/hello.txt")
	if err == nil {
		t.Fatal("hello.txt exists")
	}

	lcCommit, err := commit.GetLatest()
	if err != nil {
		t.Fatalf("GetLatest errored: %v", err)
	}

	if lcCommit.Message != "Revert \"hello\"" {
		t.Fatalf("wrong commit message: %s", lcCommit.Message)
	}
}
