package simplecache

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// SimpleCache gives read-access to the simple cache.
type SimpleCache struct {
	dir string   // cache directory
	key []string // []entry.key
}

// Open opens the cache at dir.
func Open(dir string) (*SimpleCache, error) {
	return openCache(dir)
}

// URLs returns all the URLs currently stored.
func (c SimpleCache) URLs() []string {
	return c.key
}

// OpenURL returns the Entry specified by url.
func (c SimpleCache) OpenURL(url string) (*Entry, error) {
	hash := EntryHash(url)
	return OpenEntry(hash, c.dir)
}

func openCache(dir string) (*SimpleCache, error) {
	err := checkCache(dir)
	if err != nil {
		return nil, err
	}

	name := filepath.Join(dir, "index-dir", "the-real-index")

	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return readIndex(file)
}

func checkCache(dir string) error {
	name := filepath.Clean(dir)

	info, err := os.Stat(name)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}

	file, err := os.Open(filepath.Join(name, "index"))
	if err != nil {
		return err
	}
	defer file.Close()

	index := new(fakeIndex)
	err = binary.Read(file, binary.LittleEndian, index)
	if err != nil {
		return err
	}

	if index.Magic != initialMagicNumber {
		return errors.New("index: bad magic number")
	}

	if index.Version < indexVersion {
		return errors.New("index: bad version")
	}

	return nil
}

func readIndex(file *os.File) (*SimpleCache, error) {
	index := new(indexHeader)
	err := binary.Read(file, binary.LittleEndian, index)
	if err != nil {
		return nil, err
	}

	if index.Magic != indexMagicNumber {
		return nil, errors.New("the-real-index: bad magic number")
	}
	if index.Version != indexVersion {
		return nil, errors.New("the-real-index: bad version")
	}

	dir := filepath.Dir

	cache := &SimpleCache{
		dir: dir(dir(file.Name())),
		key: make([]string, index.EntryCount),
	}

	buf := make([]byte, 8)
	offset := indexHeaderSize

	for i := uint64(0); i < index.EntryCount; i++ {
		_, err = file.ReadAt(buf, offset)
		if err != nil {
			break
		}

		hash := binary.LittleEndian.Uint64(buf)
		offset += indexEntrySize

		entry, ere := OpenEntry(hash, cache.dir)
		if ere != nil {
			log.Println(ere)
			continue
		}

		cache.key[i] = entry.URL
	}

	return cache, err
}
