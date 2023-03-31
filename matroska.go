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
	BlockTypeBlock BlockType = iota
	BlockTypeSimpleBlock
)

type BlockType uint8

const (
	BlockFlagReserved  uint8 = 0b11110000
	BlockFlagInvisible uint8 = 0b00001000
	BlockFlagLacing    uint8 = 0b00000110
	BlockFlagNotUsed   uint8 = 0b00000001

	SimpleBlockFlagKeyframe    uint  = 0b10000000
	SimpleBlockFlagReserved    uint  = 0b01110000
	SimpleBlockFlagInvisible   uint  = 0b00001000
	SimpleBlockFlagLacing      uint8 = 0b00000110
	SimpleBlockFlagDiscardable uint8 = 0b00000001
)

// Block implements block structure according to Section 6.2.3 of draft-ietf-cellar-matroska-07
// https://datatracker.ietf.org/doc/html/draft-ietf-cellar-matroska-07#section-6.2.3
type Block struct {
	Type BlockType

	TrackNumber int64
	// Relative Timestamp to Cluster timestamp, signed int16
	Timestamp time.Duration
	Duration  time.Duration
	Flags     uint8
	Data      io.Reader
}

func NewBlock(block []byte, tsoffset, scale, d time.Duration, t BlockType) (Block, error) {
	r := bytes.NewReader(block)
	tn, _, err := ebml.ReadElementDataSize(r, 8)
	if err != nil {
		return Block{}, err
	}
	tsBytes := make([]byte, 2)
	if _, err := r.Read(tsBytes); err != nil {
		return Block{}, err
	}
	ts := time.Duration(int16(binary.BigEndian.Uint16(tsBytes)))
	flag, err := r.ReadByte()
	if err != nil {
		return Block{}, err
	}
	return Block{
		Type:        t,
		TrackNumber: tn.Size(),
		Timestamp:   (tsoffset + ts) * scale,
		Duration:    d * scale,
		Flags:       flag,
		Data:        r,
	}, nil
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
