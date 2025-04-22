//go:generate go run make_doctype.go

// Package matroska contains types and structures for parsing
// matroska (.mkv, .mk3d, .mka, .mks) files.
package matroska

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"github.com/coding-socks/ebml"
	"github.com/coding-socks/ebml/ebmltext"
	"github.com/coding-socks/ebml/schema"
	"io"
	"time"
)

const DocType = "matroska"

func init() {
	var s schema.Schema
	if err := xml.Unmarshal(docType, &s); err != nil {
		panic("not able to parse matroska schema: " + err.Error())
	}
	ebml.Register(s.DocType, s)
}

const (
	BlockFlagReserved  uint8 = 0b11110000
	BlockFlagInvisible uint8 = 0b00001000
	BlockFlagLacing    uint8 = 0b00000110
	BlockFlagNotUsed   uint8 = 0b00000001

	SimpleBlockFlagKeyframe    uint8 = 0b10000000
	SimpleBlockFlagReserved    uint8 = 0b01110000
	SimpleBlockFlagInvisible   uint8 = 0b00001000
	SimpleBlockFlagLacing      uint8 = 0b00000110
	SimpleBlockFlagDiscardable uint8 = 0b00000001

	LacingFlagNo        uint8 = 0b00000000
	LacingFlagXiph      uint8 = 0b00000010
	LacingFlagEBML      uint8 = 0b00000110
	LacingFlagFixedSize uint8 = 0b00000100
)

// Frames implements block lacing according to Section 10.3 of rfc9559
// https://datatracker.ietf.org/doc/html/rfc9559#name-block-lacing
func Frames(flags uint8, data []byte) [][]byte {
	if flags == LacingFlagNo {
		return [][]byte{data}
	}
	n := data[0]
	frames := make([][]byte, int(n)+1)
	sizes := make([]uint64, int(n))
	data = data[1:]
	switch flags {
	case LacingFlagXiph:
		for i, j := 0, 0; j < len(sizes); i++ {
			sizes[j] += uint64(data[0])
			if data[0] != 0xff {
				j++
			}
			data = data[1:]
		}
	case LacingFlagEBML:
		ds, m, _ := ebmltext.ReadVintData(data)
		data = data[m:]
		sizes[0] = ds
		for i := 1; i < len(sizes); i++ {
			ds, m, _ := ebmltext.ReadVintData(data)
			s := int64(ds)
			s -= (1 << (7*m - 1)) - 1
			data = data[m:]
			sizes[i] = uint64(int64(sizes[i-1]) + s)
		}
	case LacingFlagFixedSize:
		for i := 0; i < len(sizes); i++ {
			sizes[i] = uint64(len(data) / len(frames))
		}
	}
	for i, size := range sizes {
		frames[i] = data[:size]
		data = data[size:]
	}
	frames[len(frames)-1] = data
	return frames
}

// Block implements block structure according to Section 10.1 of rfc9559
// https://datatracker.ietf.org/doc/html/rfc9559#name-block-structure
type Block struct {
	trackNumber uint
	// Relative Timestamp to Cluster timestamp, signed int16
	timestamp time.Duration
	flags     uint8
	data      []byte
}

func ReadBlock(block []byte, tsoffset time.Duration) (Block, error) {
	tn, w, err := ebmltext.ReadVintData(block[:8])
	if err != nil {
		return Block{}, err
	}
	block = block[w:] // the following elemts are located at a postion relative to the track number
	tsBytes := block[:2]
	ts := time.Duration(int16(binary.BigEndian.Uint16(tsBytes)))
	flag := block[2]
	return Block{
		trackNumber: uint(tn),
		timestamp:   tsoffset + ts,
		flags:       flag,
		data:        block[3:],
	}, nil
}

func (b Block) TrackNumber() uint {
	return b.trackNumber
}

func (b Block) Timestamp(scale time.Duration) time.Duration {
	return b.timestamp * scale
}

func (b Block) Flags() uint8 {
	return b.flags
}

func (b Block) Data() io.Reader {
	return bytes.NewReader(b.data)
}

func (b Block) Frames() [][]byte {
	return Frames(b.flags&BlockFlagLacing, b.data)
}

// SimpleBlock implements block structure according to Section 10.2 of rfc9559
// https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-21.html#name-simpleblock-structure
type SimpleBlock struct {
	trackNumber uint
	// Relative Timestamp to Cluster timestamp, signed int16
	timestamp time.Duration
	flags     uint8
	data      []byte
}

func ReadSimpleBlock(block []byte, tsoffset time.Duration) (SimpleBlock, error) {
	tn, w, err := ebmltext.ReadVintData(block[:8])
	if err != nil {
		return SimpleBlock{}, err
	}
	block = block[w:]
	tsBytes := block[:2]
	ts := time.Duration(int16(binary.BigEndian.Uint16(tsBytes)))
	flag := block[2]
	return SimpleBlock{
		trackNumber: uint(tn),
		timestamp:   tsoffset + ts,
		flags:       flag,
		data:        block[3:],
	}, nil
}

func (b SimpleBlock) TrackNumber() uint {
	return b.trackNumber
}

func (b SimpleBlock) Timestamp(scale time.Duration) time.Duration {
	return b.timestamp * scale
}

func (b SimpleBlock) Flags() uint8 {
	return b.flags
}

func (b SimpleBlock) Data() io.Reader {
	return bytes.NewReader(b.data)
}

func (b SimpleBlock) Frames() [][]byte {
	return Frames(b.flags&SimpleBlockFlagLacing, b.data)
}

func ExtractTract(w io.Writer, s *Scanner, t TrackEntry) error {
	prefix, _, _ := CodecID(t.CodecID)
	switch prefix {
	case CodecTypeVideo:
		return extractTrackVideo(w, s, t)
	case CodecTypeAudio:
		return extractTrackAudio(w, s, t)
	case CodecTypeSubtitle:
		return extractTrackSubtitle(w, s, t)
	}
	return fmt.Errorf("matroska: unknown codec %s", t.CodecID)
}
