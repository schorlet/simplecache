package simplecache

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// Cache gives read-access to the simple cache.
type Cache struct {
	dir    string // cache directory
	once   sync.Once
	hashes []uint64 // []entry.hash
	urls   []string // []entry.key
}

// Open opens the simple cache at dir.
//
// On linux, valid cache paths are:
//  ~/.cache/chromium/Default/Cache
//  ~/.cache/chromium/Default/Media Cache
func Open(dir string) (*Cache, error) {
	err := checkCache(dir)
	if err != nil {
		return nil, fmt.Errorf("invalid cache: %v", err)
	}

	name := filepath.Join(dir, "index-dir", "the-real-index")
	file, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("unable to open index: %v", err)
	}
	defer close(file)

	return readIndex(file)
}

// OpenURL returns the Entry specified by url.
func (c *Cache) OpenURL(url string) (*Entry, error) {
	hash := Hash(url)
	return OpenEntry(hash, c.dir)
}

// URLs returns all the URLs currently stored.
func (c *Cache) URLs() []string {
	c.once.Do(c.readURLs)
	return c.urls
}

func (c *Cache) readURLs() {
	c.urls = make([]string, 0, len(c.hashes))

	for _, hash := range c.hashes {
		url, err := readURL(hash, c.dir)
		if err != nil {
			log.Printf("Unable to read hash %016x: %v\n", hash, err)
			continue
		}
		c.urls = append(c.urls, url)
	}
}

func checkCache(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("unable to stat %q: %v", dir, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("not a directory: %q", dir)
	}

	file, err := os.Open(filepath.Join(dir, "index"))
	if err != nil {
		return fmt.Errorf("unable to open fakeIndex: %v", err)
	}
	defer close(file)

	index := new(fakeIndex)
	err = binary.Read(file, binary.LittleEndian, index)
	if err != nil {
		return fmt.Errorf("unable to read fakeIndex: %v", err)
	}

	if index.Magic != initialMagicNumber {
		return fmt.Errorf("bad magic number: %x, want: %x",
			index.Magic, initialMagicNumber)
	}

	if index.Version < indexVersion {
		return fmt.Errorf("bad version: %d, want: >=%d",
			index.Version, indexVersion)
	}

	return nil
}

func readIndex(file *os.File) (*Cache, error) {
	index := new(indexHeader)
	err := binary.Read(file, binary.LittleEndian, index)
	if err != nil {
		return nil, fmt.Errorf("unable to read index: %v", err)
	}

	if index.Magic != indexMagicNumber {
		return nil, fmt.Errorf("bad magic number: %x, want: %x",
			index.Magic, indexMagicNumber)
	}
	if index.Version < indexVersion {
		return nil, fmt.Errorf("bad version: %d, want: >=%d",
			index.Version, indexVersion)
	}

	dir := filepath.Dir

	cache := &Cache{
		dir:    dir(dir(file.Name())),
		hashes: make([]uint64, index.EntryCount),
	}

	if index.Version > indexVersion {
		var reasonSize int64 = 4 // last write reason
		_, err = file.Seek(reasonSize, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("unable to read 'last write reason': %v", err)
		}
	}

	entry := new(indexEntry)
	for i := uint64(0); i < index.EntryCount; i++ {
		err = binary.Read(file, binary.LittleEndian, entry)
		if err != nil {
			return nil, fmt.Errorf("unable to read entry: %v", err)
		}
		cache.hashes[i] = entry.Hash
	}

	return cache, nil
}
