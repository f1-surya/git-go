package tree

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/f1-surya/git-go/index"
	"github.com/f1-surya/git-go/object"
)

type TreeEntry struct {
	Mode uint32
	Type string
	Name string
	Hash []byte
}

type Tree struct {
	Children []TreeEntry
}

func (t *Tree) GetBlob() []byte {
	var buffer bytes.Buffer

	for _, entry := range t.Children {
		fmt.Fprintf(&buffer, "%o %s\x00", entry.Mode, entry.Name)
		buffer.Write(entry.Hash)
	}

	content := buffer.Bytes()
	var result bytes.Buffer
	fmt.Fprintf(&result, "tree %d\x00", len(content))
	result.Write(content)

	return result.Bytes()
}

func (t *Tree) Hash() [20]byte {
	return sha1.Sum(t.GetBlob())
}

// Creates the necessary trees for all the tracked files.
func CreateRoot() (map[string]*Tree, error) {
	entries, err := index.ReadIndex()
	if err != nil {
		return nil, err
	}

	trees := make(map[string]*Tree)
	trees["."] = &Tree{}

	for _, entry := range entries {
		currentTree := trees["."]
		parts := strings.Split(entry.Path, string(filepath.Separator))
		currentPath := "."

		for i := range len(parts) - 1 {
			currentPath = parts[i]
			subTree, ok := trees[currentPath]
			if !ok {
				subTree = &Tree{}
				trees[currentPath] = subTree
				currentTree.Children = append(currentTree.Children, TreeEntry{
					Mode: 0o40000,
					Type: "tree",
					Name: parts[i],
					Hash: []byte{},
				})
			}
			currentTree = subTree
		}

		fileName := parts[len(parts)-1]
		currentTree.Children = append(currentTree.Children, TreeEntry{
			Mode: entry.Mode,
			Type: "blob",
			Name: fileName,
			Hash: entry.Hash[:],
		})
	}

	for path, tree := range trees {
		if path != "." {
			hash := tree.Hash()
			parentPath := filepath.Dir(path)
			parentTree := trees[parentPath]

			for i, entry := range parentTree.Children {
				if entry.Name == filepath.Base(path) {
					parentTree.Children[i].Hash = hash[:]
					break
				}
			}
		}
	}

	return trees, nil
}

// Creates the trees for all the tracked files and writes them to the ObjectDB
func WriteTrees() (string, error) {
	trees, err := CreateRoot()
	if err != nil {
		return "", err
	}

	rootHash := trees["."].Hash()
	for _, tree := range trees {
		treeHash := tree.Hash()
		treeHashString := hex.EncodeToString(treeHash[:])
		exists := object.ObjectExist(treeHashString)
		if exists {
			continue
		}
		err = object.WriteObject(tree.GetBlob(), treeHashString)
		if err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(rootHash[:]), nil
}

// Parses the object of the given hash and returns all the children of the tree.
func ParseTreeObject(hash string) (Tree, error) {
	var root Tree

	treeObject, err := object.ReadObject(hash)
	if err != nil {
		return root, err
	}
	headerEnd := bytes.IndexByte(treeObject, '\000')
	if headerEnd == -1 {
		return root, errors.New("header end not found")
	}
	buff := bytes.NewBuffer(treeObject[headerEnd+1:])

	for buff.Len() > 0 {
		mode, err := buff.ReadBytes(' ')
		if err != nil {
			if err == io.EOF {
				break
			}
			return root, err
		}
		modeStr := string(mode[:len(mode)-1])
		var modeInt uint32
		fmt.Sscanf(modeStr, "%o", &modeInt)

		pathEnd, err := buff.ReadBytes('\000')
		if err != nil {
			return root, fmt.Errorf("missing null terminator after file path: %w", err)
		}
		path := string(pathEnd[:len(pathEnd)-1])

		hash := make([]byte, 20)
		n, err := buff.Read(hash)
		if err != nil || n != 20 {
			return root, fmt.Errorf("incomplete hash: %w", err)
		}

		entryType := "blob"
		if modeInt == object.ModeDirectory {
			entryType = "tree"
		}

		root.Children = append(root.Children, TreeEntry{
			Mode: modeInt,
			Name: path,
			Type: entryType,
			Hash: hash,
		})
	}

	return root, nil
}

// Gets the tree for the given hash and all of its subtrees recursively
func GetTreesRecursive(tree string) (map[string]Tree, error) {
	trees := make(map[string]Tree)
	var entries []TreeEntry

	root, err := ParseTreeObject(tree)
	if err != nil {
		return trees, err
	}
	entries = append(entries, root.Children...)

	for i := range entries {
		currEntry := entries[i]
		if currEntry.Type == "tree" {
			tree, err := ParseTreeObject(hex.EncodeToString(currEntry.Hash))
			if err != nil {
				return trees, err
			}
			trees[currEntry.Name] = tree
			entries = append(entries, tree.Children...)
		}
	}

	trees["."] = root

	return trees, nil
}

func GetFileHash(trees map[string]Tree, file string) string {
	currentTree, exists := trees["."]
	if !exists {
		return ""
	}

	parts := strings.Split(file, string(filepath.Separator))

	for _, part := range parts {
		for _, child := range currentTree.Children {
			if child.Type == "blob" && child.Name == part {
				return hex.EncodeToString(child.Hash)
			} else if child.Name == part {
				nextTree, ok := trees[child.Name]
				if ok {
					currentTree = nextTree
					break
				} else {
					return ""
				}
			}
		}
	}
	return ""
}

// Walks through the map and returns a map of all the files in the
// tree with the file's path as key and its hash as value
func GetAllFiles(trees map[string]Tree) map[string]string {
	files := make(map[string]string)

	var walkTrees func(string, Tree)
	walkTrees = func(prefix string, tree Tree) {
		for _, child := range tree.Children {
			path := filepath.Join(prefix, child.Name)
			if child.Type == "blob" {
				files[path] = hex.EncodeToString(child.Hash)
			} else if child.Type == "tree" {
				subTree, ok := trees[child.Name]
				if ok {
					walkTrees(path, subTree)
				}
			}
		}
	}

	root, ok := trees["."]
	if ok {
		walkTrees("", root)
	}

	return files
}
