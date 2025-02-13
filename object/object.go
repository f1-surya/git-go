package object

import (
	"bytes"
	"compress/zlib"
	"io"
	"os"
	"path/filepath"
)

func WriteObject(fileContent []byte, name string) error {
	var buffer bytes.Buffer
	w := zlib.NewWriter(&buffer)
	_, err := w.Write(fileContent)
	if err != nil {
		return err
	}
	w.Close()

	compressedContent := buffer.Bytes()

	dirPath := filepath.Join(".git-go", "objects", name[38:])

	if err = os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(dirPath, name))
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(compressedContent); err != nil {
		return err
	}

	return nil
}

func ReadObject(name string) ([]byte, error) {
	path := filepath.Join(".git-go", "objects", name[38:], name)
	compressedData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	bytesReader := bytes.NewReader(compressedData)
	reader, err := zlib.NewReader(bytesReader)
	if err != nil {
		return nil, err
	}
	reader.Close()

	decompressedContent, err := io.ReadAll(reader)
	return decompressedContent, err
}

func ObjectExist(hash string) bool {
	_, err := os.Stat(filepath.Join(".git-go", "objects", hash[38:], hash))
	return err == nil
}
