package simplecache

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// URLs returns all the URLs currently stored.
//
// On linux, valid cache paths are:
//  ~/.cache/chromium/Default/Cache
//  ~/.cache/chromium/Default/Media Cache
func URLs(path string) ([]string, error) {
	var urls []string
	if err := checkIndex(path); err != nil {
		return urls, fmt.Errorf("invalid cache: %v", err)
	}

	hashes, err := readRealIndex(path)
	if err != nil {
		return urls, fmt.Errorf("invalid cache: %v", err)
	}
	urls = make([]string, 0, len(hashes))

	for i := 0; i < len(hashes); i++ {
		url, err := readURL(hashes[i], path)
		if err != nil {
			log.Printf("Unable to read url %s: %v\n", url, err)
			continue
		}
		urls = append(urls, url)
	}
	return urls, nil
}

// checkIndex verifies the index file format.
func checkIndex(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("unable to stat %q: %v", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %q", path)
	}

	file, err := os.Open(filepath.Join(path, "index"))
	if err != nil {
		return fmt.Errorf("unable to open fakeIndex: %v", err)
	}
	defer close(file)

	var index fakeIndex
	err = binary.Read(file, binary.LittleEndian, &index)
	if err != nil {
		return fmt.Errorf("unable to read fakeIndex: %v", err)
	}

	if index.Magic != initialMagicNumber {
		return fmt.Errorf("bad magic number: %x, want: %x",
			index.Magic, initialMagicNumber)
	}
	if index.Version < indexVersion {
		return fmt.Errorf("bad version: %d, want: >= %d",
			index.Version, indexVersion)
	}
	return nil
}

// readRealIndex reads every index-entries in "the-real-index" file.
func readRealIndex(path string) ([]uint64, error) {
	var hashes []uint64

	name := filepath.Join(path, "index-dir", "the-real-index")
	file, err := os.Open(name)
	if err != nil {
		return hashes, fmt.Errorf("unable to open the-real-index: %v", err)
	}
	defer close(file)

	var index indexHeader
	err = binary.Read(file, binary.LittleEndian, &index)
	if err != nil {
		return hashes, fmt.Errorf("unable to read the-real-index: %v", err)
	}

	if err := checkRealIndex(index); err != nil {
		return hashes, fmt.Errorf("invalid header: %v", err)
	}
	if index.Version > indexVersion {
		var reasonSize int64 = 4 // last write reason
		_, err = file.Seek(reasonSize, io.SeekCurrent)
		if err != nil {
			return hashes, fmt.Errorf("unable to read 'last write reason': %v", err)
		}
	}

	hashes = make([]uint64, index.EntryCount)
	var entry indexEntry

	for i := uint64(0); i < index.EntryCount; i++ {
		err = binary.Read(file, binary.LittleEndian, &entry)
		if err != nil {
			return nil, fmt.Errorf("unable to read entry: %v", err)
		}
		hashes[i] = entry.Hash
	}

	return hashes, nil
}

// checkRealIndex verifies the "the-real-index" header.
func checkRealIndex(index indexHeader) error {
	if index.Magic != indexMagicNumber {
		return fmt.Errorf("bad magic number: %x, want: %x",
			index.Magic, indexMagicNumber)
	}
	if index.Version < indexVersion {
		return fmt.Errorf("bad version: %d, want: >= %d",
			index.Version, indexVersion)
	}
	return nil
}
