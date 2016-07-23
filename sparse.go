package simplecache

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path"
	"sort"
)

func newSparseReader(hash uint64, dir string) (io.ReadCloser, error) {
	name := path.Join(dir, fmt.Sprintf("%016x_s", hash))
	file, err := os.Open(name)
	if err != nil {
		return nil, ErrNotFound
	}

	header := new(entryHeader)
	err = binary.Read(file, binary.LittleEndian, header)
	if err != nil {
		return nil, err
	}

	if header.Magic != initialMagicNumber {
		return nil, errors.New("sparse: bad magic number")
	}
	// entryVersion ??
	if header.Version != indexVersion {
		return nil, errors.New("sparse: bad version")
	}

	reader := &sparseReader{
		file: file,
	}

	offset := entryHeaderSize + int64(header.KeyLen)
	err = reader.scan(offset)

	return reader, err
}

// sparseReader reads sparse files.
type sparseReader struct {
	file   *os.File
	ranges sparseRanges
	index  int
	stream []byte
	r, w   int64
}

func (sr sparseReader) Close() error {
	return sr.file.Close()
}

func (sr *sparseReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}

	if sr.r == sr.w {
		if err = sr.fill(); err != nil {
			return
		}
	}

	n = copy(p, sr.stream[sr.r:])
	sr.r += int64(n)

	return
}

func (sr *sparseReader) scan(offset int64) (err error) {
	for {
		_, err = sr.file.Seek(offset, 0)
		if err != nil {
			break
		}

		rangeHeader := new(sparseRangeHeader)
		err = binary.Read(sr.file, binary.LittleEndian, rangeHeader)
		if err != nil {
			break
		}

		if rangeHeader.Magic != sparseMagicNumber {
			err = errors.New("range: bad magic number")
			break
		}

		rng := sparseRange{
			Offset:     rangeHeader.Offset,
			Len:        rangeHeader.Len,
			CRC:        rangeHeader.CRC,
			FileOffset: offset + sparseRangeHeaderSize,
		}

		sr.ranges = append(sr.ranges, rng)

		offset += sparseRangeHeaderSize + rangeHeader.Len
	}

	if err != io.EOF {
		return err
	}

	sort.Sort(sr.ranges)
	return nil
}

func (sr *sparseReader) fill() error {
	if sr.index == len(sr.ranges) {
		return io.EOF
	}

	rng := sr.ranges[sr.index]

	_, err := sr.file.Seek(rng.FileOffset, 0)
	if err != nil {
		return err
	}

	sr.stream = make([]byte, rng.Len)
	_, err = io.ReadFull(sr.file, sr.stream)
	if err != nil {
		return err
	}

	actualCRC := crc32.ChecksumIEEE(sr.stream)
	if actualCRC != rng.CRC {
		return errors.New("range: bad CRC")
	}

	sr.r, sr.w = 0, rng.Len
	sr.index++

	return nil
}
