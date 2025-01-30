package utils

import (
	"crypto/sha1"
	"os"
)

// Return a hash of len 20 base on the content of the file.
func HashFile(path string) ([20]byte, uint32, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return [20]byte{}, uint32(0), err
	}
	return sha1.Sum(data), uint32(len(data)), nil
}
