package object_test

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"reflect"
	"testing"

	"github.com/f1-surya/git-go/commands"
	"github.com/f1-surya/git-go/object"
)

func TestWriteBlob(t *testing.T) {
	file, err := os.ReadFile("object_test.go")
	if err != nil {
		t.Fatalf("Read file errored: %v", err)
	}

	fileName := sha1.Sum(file)
	err = object.WriteObject(file, hex.EncodeToString(fileName[:]))
	if err != nil {
		t.Fatalf("Blob writing errored: %v", err)
	}

	os.RemoveAll(".git-go")
}

func TestContent(t *testing.T) {
	file, err := os.ReadFile("object_test.go")
	if err != nil {
		t.Fatalf("Read file errored: %v", err)
	}

	sum := sha1.Sum(file)
	fileName := hex.EncodeToString(sum[:])
	err = object.WriteObject(file, fileName)
	if object.WriteObject(file, fileName); err != nil {
		t.Fatalf("Object writing errored: %v", err)
	}

	blob, err := object.ReadObject(fileName)
	if err != nil {
		t.Fatalf("Read blob object: %v", err)
	}
	if !reflect.DeepEqual(file, blob) {
		t.Fatalf("Decompressed content doesn't match")
	}

	os.RemoveAll(".git-go")
}

func TestObjectExists(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll(".git-go")
	})
	commands.Init()
	err := commands.Add([]string{"object.go"})
	if err != nil {
		t.Fatalf("Add errored: %v", err)
	}
	file, err := os.ReadFile("object.go")
	if err != nil {
		t.Fatalf("Readfile errored: %v", err)
	}
	fileHash := sha1.Sum(file)
	fileHashString := hex.EncodeToString(fileHash[:])
	if !object.ObjectExist(fileHashString) {
		t.Fatalf("Object does exist but the function is returning false")
	}

	file, err = os.ReadFile("object_test.go")
	if err != nil {
		t.Fatalf("Read file errored: %v", err)
	}
	fileHash = sha1.Sum(file)
	fileHashString = hex.EncodeToString(fileHash[:])
	if object.ObjectExist(fileHashString) {
		t.Fatalf("File is not being tracked but the function returns true")
	}
}
