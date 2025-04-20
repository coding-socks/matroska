package ogg

import (
	"hash"
	"hash/crc32"
)

var crcTable = makeOggCRC32Table()

func makeOggCRC32Table() *crc32.Table {
	crc32.NewIEEE()
	t := new(crc32.Table)
	for i := 0; i < 256; i++ {
		r := uint32(i) << 24
		for j := 0; j < 8; j++ {
			if r&0x80000000 != 0 {
				r = (r << 1) ^ 0x04c11db7
			} else {
				r <<= 1
			}
		}
		t[i] = r
	}
	return t
}

func update(crc uint32, tab *crc32.Table, data []byte) uint32 {
	for _, v := range data {
		crc = (crc << 8) ^ tab[byte(crc>>24)^v]
	}
	return crc
}

func CRC32Checksum(data []byte) uint32 {
	return update(0, crcTable, data)
}

// digest represents the partial evaluation of a checksum.
type digest struct {
	crc uint32
	tab *crc32.Table
}

func NewCRC32() hash.Hash32 {
	return &digest{0, crcTable}
}

func (d *digest) Size() int { return crc32.Size }

func (d *digest) BlockSize() int { return 1 }

func (d *digest) Reset() { d.crc = 0 }

func (d *digest) Write(p []byte) (n int, err error) {
	// We only create digest objects through New() which takes care of
	// initialization in this case.
	d.crc = update(d.crc, d.tab, p)
	return len(p), nil
}

func (d *digest) Sum32() uint32 { return d.crc }

func (d *digest) Sum(in []byte) []byte {
	s := d.Sum32()
	return append(in, byte(s>>24), byte(s>>16), byte(s>>8), byte(s))
}
