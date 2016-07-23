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
	"path"
)

// ErrNotFound is returned when an entry is not found.
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
func OpenEntry(hash uint64, dir string) (*Entry, error) {
	name := path.Join(dir, fmt.Sprintf("%016x_0", hash))
	file, err := os.Open(name)
	if err != nil {
		return nil, ErrNotFound
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

func (e *Entry) readHeader(file *os.File) error {
	header := new(entryHeader)
	err := binary.Read(file, binary.LittleEndian, header)
	if err != nil {
		return err
	}

	if header.Magic != initialMagicNumber {
		return errors.New("entry: bad magic number")
	}
	if header.Version != entryVersion {
		return errors.New("entry: bad version")
	}

	// keyLen
	e.keyLen = int64(header.KeyLen)

	key := make([]byte, header.KeyLen)
	err = binary.Read(file, binary.LittleEndian, &key)
	if err != nil {
		return err
	}
	if header.KeyHash != superFastHash(key) {
		return errors.New("entry: bad key hash")
	}

	// URL
	e.URL = string(key)

	return nil
}

func (e *Entry) readStream0(file *os.File) error {
	stream0EOF := new(entryEOF)

	_, err := file.Seek(-1*entryEOFSize, 2)
	if err != nil {
		return err
	}

	err = binary.Read(file, binary.LittleEndian, stream0EOF)
	if err != nil {
		return err
	}

	if stream0EOF.Magic != finalMagicNumber {
		return errors.New("stream0: bad magic number")
	}

	// dataSize0
	e.dataSize0 = int64(stream0EOF.StreamSize)

	// offset0
	e.offset0 = e.fileSize - entryEOFSize - e.dataSize0
	if stream0EOF.HasSHA256() {
		e.offset0 -= int64(sha256.Size)
	}

	// verifyStream0
	if stream0EOF.Flag != 0 {
		stream0 := make([]byte, e.dataSize0)
		_, err := file.ReadAt(stream0, e.offset0)
		if err != nil {
			return err
		}

		if stream0EOF.HasCRC32() {
			actualCRC := crc32.ChecksumIEEE(stream0)
			if actualCRC != stream0EOF.CRC {
				return errors.New("stream0: bad CRC")
			}
		}

		// TODO: untested
		if stream0EOF.HasSHA256() {
			var expectedSum256 [sha256.Size]byte
			offset256 := e.offset0 + e.dataSize0
			_, err = file.ReadAt(expectedSum256[:], offset256)
			if err != nil {
				return err
			}

			actualSum256 := sha256.Sum256([]byte(e.URL))
			if actualSum256 != expectedSum256 {
				return errors.New("stream0: bad Sum256")
			}
		}
	}

	return nil
}

func (e *Entry) readStream1(file *os.File) error {
	stream1EOF := new(entryEOF)

	_, err := file.Seek(e.offset0-entryEOFSize, 0)
	if err != nil {
		return err
	}

	err = binary.Read(file, binary.LittleEndian, stream1EOF)
	if err != nil {
		return err
	}

	if stream1EOF.Magic != finalMagicNumber {
		return errors.New("stream1: bad magic number")
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
		if actualCRC != stream1EOF.CRC {
			return errors.New("stream1: bad CRC")
		}
	}

	return nil
}

// Header returns the HTTP header.
func (e Entry) Header() (http.Header, error) {
	name := path.Join(e.dir, fmt.Sprintf("%016x_0", e.hash))
	file, err := os.Open(name)
	if err != nil {
		return nil, ErrNotFound
	}
	defer file.Close()

	stream0 := make([]byte, e.dataSize0)
	_, err = file.ReadAt(stream0, e.offset0)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(stream0)

	var meta struct {
		InfoSize     int32
		Flag         int32
		RequestTime  int64
		ResponseTime int64
		HeaderSize   int32
	}
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

	name := path.Join(e.dir, fmt.Sprintf("%016x_0", e.hash))
	file, err := os.Open(name)
	if err != nil {
		return nil, ErrNotFound
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
