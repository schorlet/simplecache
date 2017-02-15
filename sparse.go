package simplecache

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sort"
)

func newSparseReader(hash uint64, dir string) (io.ReadCloser, error) {
	name := filepath.Join(dir, fmt.Sprintf("%016x_s", hash))
	file, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("unable to open entry: %v", err)
	}

	header := new(entryHeader)
	err = binary.Read(file, binary.LittleEndian, header)
	if err != nil {
		return nil, fmt.Errorf("unable to read entry header: %v", err)
	}

	if header.Magic != initialMagicNumber {
		return nil, fmt.Errorf("bad magic number: %x, want: %x",
			header.Magic, initialMagicNumber)
	}

	// entryVersion ??
	if header.Version < indexVersion {
		return nil, fmt.Errorf("bad version: %d, want: %d",
			header.Version, indexVersion)
	}

	offset := entryHeaderSize + int64(header.KeyLen)
	ranges, err := scan(file, offset)
	if err != nil {
		return nil, fmt.Errorf("unable to scan entry: %v", err)
	}

	return &sparseReader{
		file:   file,
		ranges: ranges,
	}, nil
}

// sparseReader reads sparse files.
//
// An sparse file consists of:
//	- an EntryHeader
//	- many of the following:
//		- a SparseRangeHeader
//		- a SparseRange
type sparseReader struct {
	file   *os.File
	ranges sparseRanges
	index  int
	stream []byte
	r, w   int64
}

func scan(file io.ReadSeeker, offset int64) (sparseRanges, error) {
	var ranges sparseRanges
	var err error

	for {
		_, err = file.Seek(offset, io.SeekStart)
		if err != nil {
			break
		}

		rangeHeader := new(sparseRangeHeader)
		err = binary.Read(file, binary.LittleEndian, rangeHeader)
		if err != nil {
			break
		}

		if rangeHeader.Magic != sparseMagicNumber {
			err = fmt.Errorf("bad magic number: %x, want: %x",
				rangeHeader.Magic, sparseMagicNumber)
			break
		}

		rng := sparseRange{
			Offset:     rangeHeader.Offset,
			Len:        rangeHeader.Len,
			CRC:        rangeHeader.CRC,
			FileOffset: offset + sparseRangeHeaderSize,
		}
		ranges = append(ranges, rng)

		offset += sparseRangeHeaderSize + rangeHeader.Len
	}

	if err != io.EOF {
		return nil, fmt.Errorf("unable to scan entry: %v", err)
	}

	sort.Sort(ranges)
	return ranges, nil
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

func (sr *sparseReader) fill() error {
	if sr.index == len(sr.ranges) {
		return io.EOF
	}

	rng := sr.ranges[sr.index]
	sr.stream = make([]byte, rng.Len)

	_, err := sr.file.ReadAt(sr.stream, rng.FileOffset)
	if err != nil {
		return fmt.Errorf("unable to read sparse: %v", err)
	}

	actualCRC := crc32.ChecksumIEEE(sr.stream)
	if rng.CRC != actualCRC {
		return fmt.Errorf("bad CRC: %x, want: %x",
			rng.CRC, actualCRC)
	}

	sr.r, sr.w = 0, rng.Len
	sr.index++

	return nil
}
