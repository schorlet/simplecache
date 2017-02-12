// Package simplecache provides support for reading Chromium simple cache.
// http://www.chromium.org/developers/design-documents/network-stack/disk-cache/very-simple-backend
package simplecache

import (
	"crypto/sha1"
	"encoding/binary"
	"log"
)

const (
	indexMagicNumber uint64 = 0x656e74657220796f
	indexVersion     uint32 = 6

	indexHeaderSize int64 = 36
	indexEntrySize  int64 = 24
)

const (
	initialMagicNumber uint64 = 0xfcfb6d1ba7725c30
	finalMagicNumber   uint64 = 0xf4fa6f45970d41d8
	entryVersion       uint32 = 5

	entryHeaderSize int64 = 20
	entryEOFSize    int64 = 20

	flagCRC32  uint32 = 1
	flagSHA256 uint32 = 2 // (1U << 1)
)

const (
	sparseMagicNumber     uint64 = 0xeb97bf016553676b
	sparseRangeHeaderSize int64  = 28
)

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
	Hash     uint64
	LastUsed int64
	Size     uint64
}

// Hash returns the hash of the specified url.
// The returned value may be used by OpenEntry.
func Hash(url string) uint64 {
	hash := sha1.New()

	hash.Reset()
	_, err := hash.Write([]byte(url))
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
/*
const delta = int64(11644473600000000)

func winTime(µsec int64) time.Time {
	return time.Unix(0, (µsec-delta)*1e3)
}
func fromTime(t time.Time) int64 {
	return t.UnixNano()/1e3 + delta
}
*/

func init() {
	index := new(indexHeader)
	if n := binary.Size(index); int64(n) != indexHeaderSize {
		log.Fatalf("indexHeader size error: %d, want: %d", n, indexHeaderSize)
	}

	entry := new(indexEntry)
	if n := binary.Size(entry); int64(n) != indexEntrySize {
		log.Fatalf("indexEntry size error: %d, want: %d", n, indexEntrySize)
	}

	entryHead := new(entryHeader)
	if n := binary.Size(entryHead); int64(n) != entryHeaderSize {
		log.Fatalf("entryHeader size error: %d, want: %d", n, entryHeaderSize)
	}

	entryEnd := new(entryEOF)
	if n := binary.Size(entryEnd); int64(n) != entryEOFSize {
		log.Fatalf("entryEOF size error: %d, want: %d", n, entryEOFSize)
	}

	rangeHeader := new(sparseRangeHeader)
	if n := binary.Size(rangeHeader); int64(n) != sparseRangeHeaderSize {
		log.Fatalf("sparseHeader size error: %d, want: %d", n, sparseRangeHeaderSize)
	}
}
