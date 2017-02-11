package simplecache

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// SimpleCache gives read-access to the simple cache.
type SimpleCache struct {
	dir    string // cache directory
	once   sync.Once
	hashes []uint64 // []entry.hash
	urls   []string // []entry.key
}

// Open opens the cache at dir.
func Open(dir string) (*SimpleCache, error) {
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

// Hashes returns all Entries key hash.
func (c *SimpleCache) Hashes() []uint64 {
	hashes := make([]uint64, len(c.hashes))
	copy(hashes, c.hashes)
	return hashes
}

// OpenURL returns the Entry specified by url.
// If the Entry does not exist, the error is ErrNotFound. Other errors may be returned for I/O problems.
func (c *SimpleCache) OpenURL(url string) (*Entry, error) {
	hash := EntryHash(url)
	return OpenEntry(hash, c.dir)
}

func (c *SimpleCache) readURLs() {
	c.urls = make([]string, 0, len(c.hashes))

	for _, hash := range c.hashes {
		url, err := readURL(hash, c.dir)
		if err != nil {
			log.Println(err)
			continue
		}
		c.urls = append(c.urls, url)
	}
}

// URLs returns all the URLs currently stored.
func (c *SimpleCache) URLs() []string {
	c.once.Do(c.readURLs)
	return c.urls
}

func checkCache(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("not a directory: %q", dir)
	}

	file, err := os.Open(filepath.Join(dir, "index"))
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
		return fmt.Errorf("fakeIndex: bad magic number: %x, want: %x",
			index.Magic, initialMagicNumber)
	}

	if index.Version < indexVersion {
		return fmt.Errorf("fakeIndex: bad version: %d, want: >=%d",
			index.Version, indexVersion)
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
		return nil, fmt.Errorf("index: bad magic number: %x, want: %x",
			index.Magic, indexMagicNumber)
	}
	if index.Version < indexVersion {
		return nil, fmt.Errorf("index: bad version: %d, want: >=%d",
			index.Version, indexVersion)
	}

	dir := filepath.Dir

	cache := &SimpleCache{
		dir:    dir(dir(file.Name())),
		hashes: make([]uint64, index.EntryCount),
	}

	if index.Version > indexVersion {
		var reasonSize int64 = 4 // last write reason
		_, err = file.Seek(reasonSize, os.SEEK_CUR)
		if err != nil {
			return nil, err
		}
	}

	entry := new(indexEntry)
	for i := uint64(0); i < index.EntryCount; i++ {
		err = binary.Read(file, binary.LittleEndian, entry)
		if err != nil {
			return nil, err
		}
		cache.hashes[i] = entry.Hash
	}

	return cache, err
}
