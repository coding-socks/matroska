// Package ogg implements the Ogg bitstream format based on RFC 3533.
// See: https://datatracker.ietf.org/doc/html/rfc3533
package ogg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

const Version byte = 0x00

var magicNumber = []byte("OggS")

const (
	// HeaderTypeContinuedPacket represents a page that contains data
	// of a packet continued from the previous page,
	HeaderTypeContinuedPacket byte = 1 << iota
	// HeaderTypeFirstPage represents the first page of a logical bitstream (bos).
	HeaderTypeFirstPage
	// HeaderTypeLastPage represents the last page of a logical bitstream (eos).
	HeaderTypeLastPage
)

// Builder provides a reusable buffer for writing an Ogg page.
type Builder struct{ buf *bytes.Buffer }

// NewBuilder initializes the buffer with the capture_pattern (4 Byte)
// and stream_structure_version (1 Byte) fields.
func NewBuilder() Builder {
	buf := bytes.NewBuffer(make([]byte, 26))
	return Builder{buf: buf}
}

var basePage = make([]byte, 26)

// B initialises a new page. This is a low level implementation, use it with care.
func (b Builder) B() HeaderTypeSet {
	b.buf.Reset()
	b.buf.Write(basePage)
	copy(b.buf.Bytes()[0:4], magicNumber)
	b.buf.Bytes()[4] = Version
	return HeaderTypeSet{buf: b.buf}
}

type HeaderTypeSet struct{ buf *bytes.Buffer }

// HeaderType adds a 1 Byte field that identify the specific type of this page.
func (b HeaderTypeSet) HeaderType(t byte) GranulePositionSet {
	b.buf.Bytes()[5] = t
	return GranulePositionSet{buf: b.buf}
}

type GranulePositionSet struct {
	buf *bytes.Buffer
}

// GranulePosition adds an 8 Byte field containing position information.
func (s GranulePositionSet) GranulePosition(gp uint64) BitstreamSerialNumSet {
	binary.LittleEndian.PutUint64(s.buf.Bytes()[6:14], gp)
	return BitstreamSerialNumSet{buf: s.buf}
}

type BitstreamSerialNumSet struct {
	buf *bytes.Buffer
}

// SerialNum adds a 4 Byte field containing the unique serial number
// by which the logical bitstream is identified.
func (s BitstreamSerialNumSet) SerialNum(gp int32) PageSequenceSet {
	binary.LittleEndian.PutUint32(s.buf.Bytes()[14:18], uint32(gp))
	return PageSequenceSet{buf: s.buf}
}

type PageSequenceSet struct {
	buf *bytes.Buffer
}

// PageSequence adds a 4 Byte field containing the sequence number of the page
// so the decoder can identify page loss.
func (s PageSequenceSet) PageSequence(ps uint32) SegmentsSet {
	binary.LittleEndian.PutUint32(s.buf.Bytes()[18:22], ps)
	return SegmentsSet{buf: s.buf}
}

type SegmentsSet struct {
	buf *bytes.Buffer
}

func (s SegmentsSet) Segments(segments [][]byte) Completed {
	clear(s.buf.Bytes()[22:26]) // Checksum placeholder

	pageSegments := 0
	for _, segment := range segments {
		pageSegments += (len(segment) / 255) + 1
	}
	s.buf.WriteByte(byte(pageSegments))
	for _, segment := range segments {
		l := len(segment)
		for ; l >= 255; l -= 255 {
			s.buf.WriteByte(255)
		}
		s.buf.WriteByte(byte(l))
	}
	for _, segment := range segments {
		s.buf.Write(segment)
	}
	bb := s.buf.Bytes()

	binary.LittleEndian.PutUint32(bb[22:26], CRC32Checksum(bb))

	return Completed{buf: s.buf}
}

type Completed struct{ buf *bytes.Buffer }

func (c Completed) Bytes() []byte {
	return c.buf.Bytes()
}

func (c Completed) WriteTo(w io.Writer) (n int64, err error) {
	return c.buf.WriteTo(w)
}

type Decoder struct {
	r   io.Reader
	buf bytes.Buffer

	serials map[uint32]bool
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r, serials: make(map[uint32]bool)}
}

type Page struct {
	GranulePosition uint64
	SerialNum       uint32
	SequenceNum     uint32
	FullLen         int
	SegmentTable    [][]byte
}

func (d *Decoder) DecodeStreams() (map[uint32]Page, error) {
	streams := make(map[uint32]Page)
	for {
		if _, err := io.CopyN(&d.buf, d.r, 27); err == io.EOF {
			return nil, io.ErrUnexpectedEOF
		} else if err != nil {
			return nil, err
		}
		b := d.buf.Bytes()
		if !bytes.Equal(b[0:4], magicNumber) {
			return nil, fmt.Errorf("ogg: invalid magic number: %s", b[0:4])
		}
		if b[4] != Version {
			return nil, fmt.Errorf("ogg: invalid version: %s", b[4:4])
		}
		if b[5] != HeaderTypeFirstPage {
			break
		}
		var p Page
		p.GranulePosition = binary.LittleEndian.Uint64(b[6:14])
		p.SerialNum = binary.LittleEndian.Uint32(b[14:18])
		if d.serials[p.SerialNum] {
			return nil, fmt.Errorf("ogg: serial number already in use: %d", p.SerialNum)
		}
		p.SequenceNum = binary.LittleEndian.Uint32(b[18:22])
		checksum := binary.LittleEndian.Uint32(b[22:26])
		numSegments := b[26]

		if _, err := io.CopyN(&d.buf, d.r, int64(numSegments)); err == io.EOF {
			return nil, io.ErrUnexpectedEOF
		} else if err != nil {
			return nil, err
		}

		fullSize := 0
		sizes := make([]uint16, 0)
		b = d.buf.Bytes()[27:]
		for i, size := byte(0), uint16(0); i < numSegments; i++ {
			fullSize += int(b[i])
			size += uint16(b[i])
			if b[i] != 255 {
				sizes = append(sizes, size)
				size = 0
			}
		}

		if _, err := io.CopyN(&d.buf, d.r, int64(fullSize)); err == io.EOF {
			return nil, io.ErrUnexpectedEOF
		} else if err != nil {
			return nil, err
		}

		b = d.buf.Bytes()[27+numSegments:]
		segments := make([]byte, fullSize)
		copy(segments, b) // make sure we don't lose the data when the buffer is overwritten
		p.SegmentTable = make([][]byte, len(sizes))
		for i := 0; i < len(sizes); i++ {
			size := sizes[i]
			p.SegmentTable[i] = b[:size]
			b = b[size:]
		}

		b = d.buf.Bytes()
		clear(b[22:26])
		if CRC32Checksum(b) != checksum {
			fmt.Printf("ogg: invalid CRC32 checksum: %x\n", checksum)
		}
		d.buf.Reset()
		streams[p.SerialNum] = p
		d.serials[p.SerialNum] = true
	}
	return streams, nil
}

func (d *Decoder) DecodePage() (*Page, error) {
	if len(d.serials) == 0 {
		return nil, io.EOF
	}
	b := d.buf.Bytes()
	if len(b) == 0 {
		if _, err := io.CopyN(&d.buf, d.r, 27); err == io.EOF {
			return nil, io.ErrUnexpectedEOF
		} else if err != nil {
			return nil, err
		}
		b = d.buf.Bytes()
		if !bytes.Equal(b[0:4], magicNumber) {
			return nil, fmt.Errorf("ogg: invalid magic number: %s", b[0:4])
		}
		if b[4] != Version {
			return nil, fmt.Errorf("ogg: invalid version: %s", b[4:4])
		}
		if b[5] == HeaderTypeFirstPage {
			return nil, fmt.Errorf("ogg: invalid header type: %d", b[5])
		}
	}
	var p Page
	p.GranulePosition = binary.LittleEndian.Uint64(b[6:14])
	p.SerialNum = binary.LittleEndian.Uint32(b[14:18])
	if !d.serials[p.SerialNum] {
		return nil, fmt.Errorf("ogg: uninitialised serial number: %d", p.SerialNum)
	}
	p.SequenceNum = binary.LittleEndian.Uint32(b[18:22])
	checksum := binary.LittleEndian.Uint32(b[22:26])
	numSegments := b[26]

	if _, err := io.CopyN(&d.buf, d.r, int64(numSegments)); err == io.EOF {
		return nil, io.ErrUnexpectedEOF
	} else if err != nil {
		return nil, err
	}

	fullSize := 0
	sizes := make([]uint16, 0)
	b = d.buf.Bytes()[27:]
	for i, size := byte(0), uint16(0); i < numSegments; i++ {
		fullSize += int(b[i])
		size += uint16(b[i])
		if b[i] != 255 {
			sizes = append(sizes, size)
			size = 0
		}
	}

	if _, err := io.CopyN(&d.buf, d.r, int64(fullSize)); err == io.EOF {
		return nil, io.ErrUnexpectedEOF
	} else if err != nil {
		return nil, err
	}

	b = d.buf.Bytes()[27+numSegments:]
	p.FullLen = fullSize
	segments := make([]byte, fullSize)
	copy(segments, b) // make sure we don't lose the data when the buffer is overwritten
	p.SegmentTable = make([][]byte, len(sizes))
	for i := 0; i < len(sizes); i++ {
		size := sizes[i]
		p.SegmentTable[i] = b[:size]
		b = b[size:]
	}

	b = d.buf.Bytes()
	clear(b[22:26])
	if CRC32Checksum(b) != checksum {
		fmt.Printf("ogg: invalid CRC32 checksum: %x\n", checksum)
	}
	if b[5] == HeaderTypeLastPage {
		delete(d.serials, p.SerialNum)
	}
	d.buf.Reset()
	return &p, nil
}
