package src

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

// GitObject - Git object type with kind, size, and data
// kind is the type of object - commit / tree / blob etc.
// size is total size of uncompressed data
// data is the data, different structure for every type
type GitObject struct {
	Kind string
	Size string
	Data []byte
}

// Write - writes compressed object files by calculating hashes as the filename (sha1)
func (obj GitObject) Write(gitdir string) (string, error) {
	// GitObject kind, size, and data in []byte
	bKind := []byte(obj.Kind)
	bSize := []byte(strconv.Itoa(len([]byte(obj.Data)) - 1)) // -1 because it didn't match the experimental result
	bData := []byte(obj.Data)

	// Attatching the chunks
	sl := [][]byte{bKind, []byte(" "), bSize, []byte{0x00}, bData}
	content := bytes.Join(sl, []byte(""))

	// Computing hash for the content
	h := sha1.New()
	h.Write(content)
	sha := h.Sum(nil)

	// String representation of "sha" (uint8) (encoded in base16)
	shaStr := hex.EncodeToString(sha)

	// Creating directory for with first two values of "shaStr"
	err := os.MkdirAll(path.Join(gitdir, "objects", shaStr[:2]), 0777)
	if err != nil {
		return "", err
	}

	// Absolute Path to the written file
	nFilePath, err := filepath.Abs(path.Join(gitdir, "objects", shaStr[:2], shaStr[2:]))
	if err != nil {
		return "", err
	}

	// Creating new file with the content
	nFile, err := os.Create(nFilePath)
	if err != nil {
		return "", err
	}
	defer nFile.Close()

	// Compressing content (zlib)
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(content)
	w.Close()

	// Writing compressed content in "nFile", the newly created file
	_, err = nFile.Write(b.Bytes())
	if err != nil {
		return "", err
	}

	// return nFilePath, nil
	return shaStr, nil
}

// ReadObjectFile - Reads the object file compressed data, returns uncompressed content
func ReadObjectFile(objectpath string) ([]byte, error) {
	// Reading file data (Compressed)
	data, err := ioutil.ReadFile(objectpath)
	if err != nil {
		return nil, err
	}

	// Decompressing data (zlib)
	b := bytes.NewReader(data)
	raw, err := zlib.NewReader(b) // req - Decompressed Data
	if err != nil {
		return nil, err
	}

	defer raw.Close()

	// Data in []byte
	bData, err := ioutil.ReadAll(raw)
	if err != nil {
		return nil, err
	}

	return bData, nil
}

// ReadObject - Reads a object and returns appropriate "GitObject" struct
func ReadObject(objectpath string) (GitObject, error) {
	// Check if file exist or not
	if _, err := os.Stat(objectpath); os.IsNotExist(err) {
		return GitObject{}, errors.New("Specifies file does not exist")
	}

	// Reading File Data
	fData, err := ReadObjectFile(objectpath)
	if err != nil {
		return GitObject{}, err
	}

	// fmt.Printf("content:\n%+s\n", fData)

	x := bytes.IndexByte(fData, byte(' '))  // Index of ' ' (rune) in file data
	y := bytes.IndexByte(fData, byte(0x00)) // Index of 0x00 (null seperator) in file data

	return GitObject{
		Kind: string(fData[:x]),
		Size: string(fData[x+1 : y]),
		Data: fData[y+1:],
	}, nil

}
