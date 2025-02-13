package tree

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/f1-surya/git-go/index"
	"github.com/f1-surya/git-go/object"
)

type TreeEntry struct {
	Mode uint32
	Type string
	Name string
	Hash [20]byte
}

type Tree struct {
	Children []TreeEntry
}

func (t *Tree) GetBlob() []byte {
	var data []byte

	for _, entry := range t.Children {
		modeBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(modeBytes, entry.Mode)
		data = append(data, modeBytes...)
		data = append(data, []byte(" "+entry.Name+"\000")...)
		data = append(data, entry.Hash[:]...)
	}

	header := fmt.Sprintf("tree %d\000", len(data))
	data = append([]byte(header), data...)

	return data
}

func (t *Tree) Hash() [20]byte {
	return sha1.Sum(t.GetBlob())
}

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

		for i := 0; i < len(parts)-1; i++ {
			currentPath = parts[i]
			subTree, ok := trees[currentPath]
			if !ok {
				subTree = &Tree{}
				trees[currentPath] = subTree
				currentTree.Children = append(currentTree.Children, TreeEntry{
					Mode: 0o40000,
					Type: "tree",
					Name: currentPath,
					Hash: [20]byte{},
				})
			}
			currentTree = subTree
		}

		fileName := parts[len(parts)-1]
		currentTree.Children = append(currentTree.Children, TreeEntry{
			Mode: entry.Mode,
			Type: "blob",
			Name: fileName,
			Hash: entry.Hash,
		})
	}

	for path, tree := range trees {
		if path != "." {
			hash := tree.Hash()
			parentPath := filepath.Dir(path)
			parentTree := trees[parentPath]

			for i, entry := range parentTree.Children {
				if entry.Name == filepath.Base(parentPath) {
					parentTree.Children[i].Hash = hash
					break
				}
			}
		}
	}

	return trees, nil
}

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
