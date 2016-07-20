package simplecache

// SuperFastHash implementation for golang:
// http://www.azillionmonkeys.com/qed/hash.html

import "encoding/binary"

func get16bits(p []byte) uint32 {
	return uint32(binary.LittleEndian.Uint16(p))
}

func superFastHash(data []byte) uint32 {
	lend := len(data)
	if lend == 0 {
		return 0
	}

	hash := uint32(lend)
	rem := lend & 3
	i := 0

	/* Main loop */
	for ; i < lend-rem; i += 4 {
		hash += get16bits(data[i : i+2])
		tmp := (get16bits(data[i+2:i+4]) << 11) ^ hash
		hash = (hash << 16) ^ tmp
		hash += hash >> 11
	}

	/* Handle end cases */
	switch rem {
	case 3:
		hash += get16bits(data[i : i+2])
		hash ^= hash << 16
		hash ^= uint32(data[i+2]) << 18
		hash += hash >> 11
	case 2:
		hash += get16bits(data[i : i+2])
		hash ^= hash << 11
		hash += hash >> 17
	case 1:
		hash += uint32(data[i])
		hash ^= hash << 10
		hash += hash >> 1
	}

	/* Force "avalanching" of final 127 bits */
	hash ^= hash << 3
	hash += hash >> 5
	hash ^= hash << 4
	hash += hash >> 17
	hash ^= hash << 25
	hash += hash >> 6

	return hash
}
