package simplecache

import (
	"encoding/binary"
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

// private: ---------------------------------

func openCache(dir string) (*SimpleCache, error) {
	err := checkCache(dir)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filepath.Join(dir, "index-dir", "the-real-index"))
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
		return fmt.Errorf("cdc: not a directory: %s", dir)
	}

	_, err = os.Stat(filepath.Join(name, "index"))
	return err
}

func readIndex(file *os.File) (*SimpleCache, error) {
	index := new(indexHeader)
	err := binary.Read(file, binary.LittleEndian, index)
	if err != nil {
		return nil, err
	}
	// fmt.Println(index)

	if index.Magic != indexMagicNumber {
		log.Fatal("bad MagicNumber")
	}

	if index.Version >= 7 {
		var reason uint32
		err := binary.Read(file, binary.LittleEndian, &reason)
		if err != nil {
			return nil, err
		}
		// fmt.Printf("Reason:%d\n", reason)
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

	// var modified int64
	// err = binary.Read(file, binary.LittleEndian, &modified)
	// if err != nil {
	// log.Fatal(err)
	// }
	// fmt.Printf("Modified:%s\n", timeFormat(winTime(modified)))

	return cache, err
}
