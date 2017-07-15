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
//
// URLs reads the files named "path/index" and "path/index-dir/the-real-index"
// and every files named "path/hash" where hash is read from the-real-index file.
//
// An error is returned on a problem with any of the two index files
// but not for each individual entries.
// In this situation the entry's URL is just not added to the returned URLs.
func URLs(path string) ([]string, error) {
	var urls []string
	if err := checkFakeIndex(path); err != nil {
		return urls, fmt.Errorf("get urls from %s: %v", path, err)
	}

	hashes, err := readRealIndex(path)
	if err != nil {
		return urls, fmt.Errorf("get urls from %s: %v", path, err)
	}
	urls = make([]string, 0, len(hashes))

	for i := 0; i < len(hashes); i++ {
		url, err := readURL(hashes[i], path)
		if err != nil {
			log.Printf("Unable to get %s from %s: %v\n", url, path, err)
			continue
		}
		urls = append(urls, url)
	}
	return urls, nil
}

// checkFakeIndex verifies the index file format.
func checkFakeIndex(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("fake-index: %v", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("fake-index: not a directory")
	}

	file, err := os.Open(filepath.Join(path, "index"))
	if err != nil {
		return fmt.Errorf("open fake-index: %v", err)
	}
	defer close(file)

	var index fakeIndex
	err = binary.Read(file, binary.LittleEndian, &index)
	if err != nil {
		return fmt.Errorf("read fake-index: %v", err)
	}

	if index.Magic != initialMagicNumber {
		return fmt.Errorf("fake-index magic: %x, want: %x",
			index.Magic, initialMagicNumber)
	}
	if index.Version < indexVersion {
		return fmt.Errorf("fake-index version: %d, want: >= %d",
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
		return hashes, fmt.Errorf("open real-index: %v", err)
	}
	defer close(file)

	var index indexHeader
	err = binary.Read(file, binary.LittleEndian, &index)
	if err != nil {
		return hashes, fmt.Errorf("read real-index header: %v", err)
	}

	if err := checkRealIndex(index); err != nil {
		return hashes, fmt.Errorf("check real-index header: %v", err)
	}
	if index.Version > indexVersion {
		var reasonSize int64 = 4 // last write reason
		_, err = file.Seek(reasonSize, io.SeekCurrent)
		if err != nil {
			return hashes, fmt.Errorf("read real-index 'last write reason': %v", err)
		}
	}

	hashes = make([]uint64, index.EntryCount)
	var entry indexEntry

	for i := uint64(0); i < index.EntryCount; i++ {
		err = binary.Read(file, binary.LittleEndian, &entry)
		if err != nil {
			return nil, fmt.Errorf("read real-index entry: %v", err)
		}
		hashes[i] = entry.Hash
	}

	return hashes, nil
}

// checkRealIndex verifies the "the-real-index" header.
func checkRealIndex(index indexHeader) error {
	if index.Magic != indexMagicNumber {
		return fmt.Errorf("magic: %x, want: %x",
			index.Magic, indexMagicNumber)
	}
	if index.Version < indexVersion {
		return fmt.Errorf("version: %d, want: >= %d",
			index.Version, indexVersion)
	}
	return nil
}
