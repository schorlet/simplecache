// Package simplecache provides support for reading Chromium simple cache.
// http://www.chromium.org/developers/design-documents/network-stack/disk-cache/very-simple-backend
package simplecache

import (
	"crypto/sha1"
	"encoding/binary"
	"log"
	"time"
)

const indexMagicNumber uint64 = 0x656e74657220796f
const indexVersion uint32 = 6

const indexHeaderSize int64 = 36
const indexEntrySize int64 = 24

const initialMagicNumber uint64 = 0xfcfb6d1ba7725c30
const finalMagicNumber uint64 = 0xf4fa6f45970d41d8
const entryVersion uint32 = 5

const entryHeaderSize int64 = 20
const entryEOFSize int64 = 20

const flagCRC32 uint32 = 1
const flagSHA256 uint32 = 2 // (1U << 1)

const sparseMagicNumber uint64 = 0xeb97bf016553676b
const sparseRangeHeaderSize int64 = 28

// fakeIndex is the content of the index file.
type fakeIndex struct {
	Magic   uint64
	Version uint32
	_       uint64
}

// indexHeader is the header of the the-real-index file.
type indexHeader struct {
	Payload    uint32
	CRC        uint32
	Magic      uint64
	Version    uint32
	EntryCount uint64
	CacheSize  uint64
}

// indexEntry is an entry in the the-real-index file.
type indexEntry struct {
	Hash      uint64
	LastUsed  int64
	EntrySize uint64
}

// EntryHash returns the hash of the specified key.
func EntryHash(key string) uint64 {
	hash := sha1.New()

	hash.Reset()
	_, err := hash.Write([]byte(key))
	if err != nil {
		return 0
	}

	// sum is [20]byte
	sum := hash.Sum(nil)

	// uses the top 64 bits
	return binary.LittleEndian.Uint64(sum[:8])
}

// entryHeader is the header of an entry file.
type entryHeader struct {
	Magic   uint64
	Version uint32
	KeyLen  int32
	KeyHash uint32
}

// entryEOF ends a stream in an entry file.
type entryEOF struct {
	Magic      uint64
	Flag       uint32
	CRC        uint32
	StreamSize int32
}

// HasCRC32
func (e entryEOF) HasCRC32() bool {
	return e.Flag&flagCRC32 != 0
}

// HasSHA256
func (e entryEOF) HasSHA256() bool {
	return e.Flag&flagSHA256 != 0
}

// sparseRangeHeader is the header of a stream range in a sparse file.
type sparseRangeHeader struct {
	Magic  uint64
	Offset int64
	Len    int64
	CRC    uint32
}

// sparseRange is a stream range in a sparse file.
type sparseRange struct {
	Offset     int64
	Len        int64
	CRC        uint32
	FileOffset int64
}

type sparseRanges []sparseRange

func (ranges sparseRanges) Len() int {
	return len(ranges)
}
func (ranges sparseRanges) Swap(i, j int) {
	ranges[i], ranges[j] = ranges[j], ranges[i]
}
func (ranges sparseRanges) Less(i, j int) bool {
	var rng0, rng1 = ranges[i], ranges[j]
	return rng0.Offset < rng1.Offset
}

// unix epoch - win epoch (µsec)
// (1970-01-01 - 1601-01-01)
const delta = int64(11644473600000000)

func winTime(µsec int64) time.Time {
	return time.Unix(0, (µsec-delta)*1e3)
}

func init() {
	index := new(indexHeader)
	if n := binary.Size(index); int64(n) != indexHeaderSize {
		log.Fatal("indexHeader size error:", n)
	}

	entry := new(indexEntry)
	if n := binary.Size(entry); int64(n) != indexEntrySize {
		log.Fatal("indexEntry size error:", n)
	}

	entryHead := new(entryHeader)
	if n := binary.Size(entryHead); int64(n) != entryHeaderSize {
		log.Fatal("entryHeader size error:", n)
	}

	entryEnd := new(entryEOF)
	if n := binary.Size(entryEnd); int64(n) != entryEOFSize {
		log.Fatal("entryEOF size error:", n)
	}

	rangeHeader := new(sparseRangeHeader)
	if n := binary.Size(rangeHeader); int64(n) != sparseRangeHeaderSize {
		log.Fatal("sparseHeader size error:", n)
	}
}
