package simplecache

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// Entry represents a HTTP response as stored in the cache.
// Each entry is stored in a file named "path/hash(url)_0".
type Entry struct {
	URL       string
	hash      uint64
	path      string
	fileSize  int64
	keyLen    int64
	offset1   int64
	dataSize1 int64
	offset0   int64
	dataSize0 int64
}

// Get returns the Entry for the specified URL.
// An error is returned if the format of the entry does not match the one expected.
func Get(url, path string) (*Entry, error) {
	sum := sha1.Sum([]byte(url))
	hash := binary.LittleEndian.Uint64(sum[:8])

	name := filepath.Join(path, fmt.Sprintf("%016x_0", hash))
	file, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("getting %s: %v", url, err)
	}
	defer close(file)

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("getting %s: %v", url, err)
	}

	entry := Entry{
		hash:     hash,
		path:     path,
		fileSize: stat.Size(),
	}

	if err = entry.readHeader(file); err != nil {
		return nil, fmt.Errorf("getting %s: %v", url, err)
	}
	if err = entry.readStream0(file); err != nil {
		return nil, fmt.Errorf("getting %s: %v", url, err)
	}
	if err = entry.readStream1(file); err != nil {
		return nil, fmt.Errorf("getting %s: %v", url, err)
	}

	return &entry, nil
}

func readURL(hash uint64, path string) (string, error) {
	name := filepath.Join(path, fmt.Sprintf("%016x_0", hash))
	file, err := os.Open(name)
	if err != nil {
		return "", fmt.Errorf("readurl %016x_0: %v", hash, err)
	}
	defer close(file)

	var entry Entry
	err = entry.readHeader(file)
	if err != nil {
		return "", fmt.Errorf("readurl %016x_0: %v", hash, err)
	}
	return entry.URL, nil
}

func (e *Entry) readHeader(file io.Reader) error {
	var header entryHeader
	err := binary.Read(file, binary.LittleEndian, &header)
	if err != nil {
		return fmt.Errorf("read header: %v", err)
	}

	if header.Magic != initialMagicNumber {
		return fmt.Errorf("header magic: %x, want: %x",
			header.Magic, initialMagicNumber)
	}
	if header.Version != entryVersion {
		return fmt.Errorf("header version: %d, want: %d",
			header.Version, entryVersion)
	}

	// keyLen
	e.keyLen = int64(header.KeyLen)

	key := make([]byte, header.KeyLen)
	err = binary.Read(file, binary.LittleEndian, &key)
	if err != nil {
		return fmt.Errorf("read header key: %v", err)
	}

	sfh := superFastHash(key)
	if header.KeyHash != sfh {
		return fmt.Errorf("header key hash: %x, want: %x",
			header.KeyHash, sfh)
	}

	// URL
	e.URL = string(key)

	return nil
}

func (e *Entry) readStream0(file *os.File) error {
	var stream0EOF entryEOF

	_, err := file.Seek(-1*entryEOFSize, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("seek stream0: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &stream0EOF)
	if err != nil {
		return fmt.Errorf("read stream0 entryEOF: %v", err)
	}

	if stream0EOF.Magic != finalMagicNumber {
		return fmt.Errorf("stream0 magic: %x, want: %x",
			stream0EOF.Magic, finalMagicNumber)
	}

	// dataSize0
	e.dataSize0 = int64(stream0EOF.StreamSize)

	// offset0
	e.offset0 = e.fileSize - entryEOFSize - e.dataSize0
	if stream0EOF.HasSHA256() {
		e.offset0 -= int64(sha256.Size)
	}

	// verifyStream0

	if stream0EOF.HasCRC32() {
		stream0 := make([]byte, e.dataSize0)
		_, err = file.ReadAt(stream0, e.offset0)
		if err != nil {
			return fmt.Errorf("read stream0: %v", err)
		}

		actualCRC := crc32.ChecksumIEEE(stream0)
		if stream0EOF.CRC != actualCRC {
			return fmt.Errorf("stream0 CRC: %x, want: %x",
				stream0EOF.CRC, actualCRC)
		}
	}

	if stream0EOF.HasSHA256() {
		var expectedSum256 [sha256.Size]byte
		offset256 := e.offset0 + e.dataSize0

		_, err = file.ReadAt(expectedSum256[:], offset256)
		if err != nil {
			return fmt.Errorf("read stream0 sha256: %v", err)
		}

		actualSum256 := sha256.Sum256([]byte(e.URL))
		if expectedSum256 != actualSum256 {
			return fmt.Errorf("stream0 sha256: %x, want: %x",
				expectedSum256, actualSum256)
		}
	}

	return nil
}

func (e *Entry) readStream1(file *os.File) error {
	var stream1EOF entryEOF

	_, err := file.Seek(e.offset0-entryEOFSize, io.SeekStart)
	if err != nil {
		return fmt.Errorf("seek stream1: %v", err)
	}

	err = binary.Read(file, binary.LittleEndian, &stream1EOF)
	if err != nil {
		return fmt.Errorf("read stream1 entryEOF: %v", err)
	}

	if stream1EOF.Magic != finalMagicNumber {
		return fmt.Errorf("stream1 magic: %x, want: %x",
			stream1EOF.Magic, finalMagicNumber)
	}

	// dataSize1
	e.dataSize1 = int64(stream1EOF.StreamSize)

	// offset1
	e.offset1 = entryHeaderSize + e.keyLen

	// verifyStream1
	if e.dataSize1 > 0 && stream1EOF.HasCRC32() {
		stream1 := make([]byte, e.dataSize1)
		_, err := file.ReadAt(stream1, e.offset1)
		if err != nil {
			return fmt.Errorf("read stream1: %v", err)
		}

		actualCRC := crc32.ChecksumIEEE(stream1)
		if stream1EOF.CRC != actualCRC {
			return fmt.Errorf("stream1 CRC: %x, want: %x",
				stream1EOF.CRC, actualCRC)
		}
	}

	return nil
}

// Header returns the HTTP header.
func (e *Entry) Header() (http.Header, error) {
	name := filepath.Join(e.path, fmt.Sprintf("%016x_0", e.hash))
	file, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open header: %v", err)
	}
	defer close(file)

	stream0 := make([]byte, e.dataSize0)
	_, err = file.ReadAt(stream0, e.offset0)
	if err != nil {
		return nil, fmt.Errorf("read header: %v", err)
	}

	var meta struct {
		InfoSize     int32
		Flag         int32
		RequestTime  int64
		ResponseTime int64
		HeaderSize   int32
	}

	reader := bytes.NewReader(stream0)
	err = binary.Read(reader, binary.LittleEndian, &meta)
	if err != nil {
		return nil, fmt.Errorf("read header metadata: %v", err)
	}

	buf := make([]byte, meta.HeaderSize)
	err = binary.Read(reader, binary.LittleEndian, buf)
	if err != nil {
		return nil, fmt.Errorf("read header size: %v", err)
	}

	header := make(http.Header)
	lines := bytes.Split(buf, []byte{0})

	for _, line := range lines {
		kv := bytes.SplitN(line, []byte{':'}, 2)
		if len(kv) == 2 {
			header.Add(
				string(bytes.TrimSpace(kv[0])),
				string(bytes.TrimSpace(kv[1])))
		}
	}

	return header, nil
}

// Body returns the HTTP body.
// Body may read a file named "path/hash(url)_s".
func (e *Entry) Body() (io.ReadCloser, error) {
	if e.dataSize1 == 0 {
		return newSparseReader(e.hash, e.path)
	}

	name := filepath.Join(e.path, fmt.Sprintf("%016x_0", e.hash))
	file, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open body: %v", err)
	}
	defer close(file)

	stream1 := make([]byte, e.dataSize1)
	_, err = file.ReadAt(stream1, e.offset1)
	if err != nil {
		return nil, fmt.Errorf("read body: %v", err)
	}

	reader := bytes.NewReader(stream1)
	return ioutil.NopCloser(reader), nil
}

func close(f *os.File) {
	if err := f.Close(); err != nil {
		log.Printf("Error closing file %s: %v\n", f.Name(), err)
	}
}
