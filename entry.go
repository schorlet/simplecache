package simplecache

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

// ErrNotFound is returned when an entry does not exist.
var ErrNotFound = errors.New("entry not found")

// Entry represents an entry as stored in the cache.
//
// An entry containing stream 0 and stream 1 in the cache consists of:
//	- a SimpleFileHeader.
//	- the key.
//	- the data from stream 1.
//	- a SimpleFileEOF record for stream 1.
//	- the data from stream 0.
//	- (optionally) the SHA256 of the key.
//	- a SimpleFileEOF record for stream 0.
type Entry struct {
	URL       string
	hash      uint64
	dir       string
	fileSize  int64
	keyLen    int64
	offset1   int64
	dataSize1 int64
	offset0   int64
	dataSize0 int64
}

// OpenEntry returns the Entry specified by hash, in the cache at dir.
// If the Entry does not exist, the error is ErrNotFound. Other errors may be returned for I/O problems.
func OpenEntry(hash uint64, dir string) (*Entry, error) {
	name := filepath.Join(dir, fmt.Sprintf("%016x_0", hash))
	file, err := os.Open(name)

	if os.IsNotExist(err) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	entry := &Entry{
		hash:     hash,
		dir:      dir,
		fileSize: stat.Size(),
	}

	err = entry.readHeader(file)
	if err != nil {
		return nil, err
	}

	err = entry.readStream0(file)
	if err != nil {
		return nil, err
	}

	err = entry.readStream1(file)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func readURL(hash uint64, dir string) (string, error) {
	name := filepath.Join(dir, fmt.Sprintf("%016x_0", hash))
	file, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer file.Close()

	entry := new(Entry)
	err = entry.readHeader(file)
	if err != nil {
		return "", err
	}
	return entry.URL, nil
}

func (e *Entry) readHeader(file *os.File) error {
	header := new(entryHeader)
	err := binary.Read(file, binary.LittleEndian, header)
	if err != nil {
		return err
	}

	if header.Magic != initialMagicNumber {
		return fmt.Errorf("entry: bad magic number: %x, want: %x",
			header.Magic, initialMagicNumber)
	}
	if header.Version != entryVersion {
		return fmt.Errorf("entry: bad version: %d, want: %d",
			header.Version, entryVersion)
	}

	// keyLen
	e.keyLen = int64(header.KeyLen)

	key := make([]byte, header.KeyLen)
	err = binary.Read(file, binary.LittleEndian, &key)
	if err != nil {
		return err
	}

	sfh := superFastHash(key)
	if header.KeyHash != sfh {
		return fmt.Errorf("entry: bad key hash: %x, want: %x",
			header.KeyHash, sfh)
	}

	// URL
	e.URL = string(key)

	return nil
}

func (e *Entry) readStream0(file *os.File) error {
	stream0EOF := new(entryEOF)

	_, err := file.Seek(-1*entryEOFSize, os.SEEK_END)
	if err != nil {
		return err
	}

	err = binary.Read(file, binary.LittleEndian, stream0EOF)
	if err != nil {
		return err
	}

	if stream0EOF.Magic != finalMagicNumber {
		return fmt.Errorf("stream0: bad magic number: %x, want: %x",
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
			return err
		}

		actualCRC := crc32.ChecksumIEEE(stream0)
		if stream0EOF.CRC != actualCRC {
			return fmt.Errorf("stream0: bad CRC: %x, want: %x",
				stream0EOF.CRC, actualCRC)
		}
	}

	if stream0EOF.HasSHA256() {
		var expectedSum256 [sha256.Size]byte
		offset256 := e.offset0 + e.dataSize0

		_, err = file.ReadAt(expectedSum256[:], offset256)
		if err != nil {
			return err
		}

		actualSum256 := sha256.Sum256([]byte(e.URL))
		if expectedSum256 != actualSum256 {
			return fmt.Errorf("stream0: bad Sum256: %x, want: %x",
				expectedSum256, actualSum256)
		}
	}

	return nil
}

func (e *Entry) readStream1(file *os.File) error {
	stream1EOF := new(entryEOF)

	_, err := file.Seek(e.offset0-entryEOFSize, os.SEEK_SET)
	if err != nil {
		return err
	}

	err = binary.Read(file, binary.LittleEndian, stream1EOF)
	if err != nil {
		return err
	}

	if stream1EOF.Magic != finalMagicNumber {
		return fmt.Errorf("stream1: bad magic number: %x, want: %x",
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
			return err
		}

		actualCRC := crc32.ChecksumIEEE(stream1)
		if stream1EOF.CRC != actualCRC {
			return fmt.Errorf("stream1: bad CRC: %x, want: %x",
				stream1EOF.CRC, actualCRC)
		}
	}

	return nil
}

// Header returns the HTTP header.
func (e Entry) Header() (http.Header, error) {
	name := filepath.Join(e.dir, fmt.Sprintf("%016x_0", e.hash))
	file, err := os.Open(name)

	if os.IsNotExist(err) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	defer file.Close()

	stream0 := make([]byte, e.dataSize0)
	_, err = file.ReadAt(stream0, e.offset0)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	buf := make([]byte, meta.HeaderSize)
	err = binary.Read(reader, binary.LittleEndian, buf)
	if err != nil {
		return nil, err
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
func (e Entry) Body() (io.ReadCloser, error) {
	if e.dataSize1 == 0 {
		return newSparseReader(e.hash, e.dir)
	}

	name := filepath.Join(e.dir, fmt.Sprintf("%016x_0", e.hash))
	file, err := os.Open(name)

	if os.IsNotExist(err) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	defer file.Close()

	stream1 := make([]byte, e.dataSize1)
	_, err = file.ReadAt(stream1, e.offset1)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(stream1)
	return ioutil.NopCloser(reader), nil
}
