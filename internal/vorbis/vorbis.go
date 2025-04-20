package vorbis

import (
	"encoding/binary"
	"fmt"
	"github.com/coding-socks/matroska/internal/ogg"
	"io"
	"math/rand/v2"
)

// IdentificationHeader is based on Section 4.2.2 of Vorbis I specification.
// See: https://xiph.org/vorbis/doc/Vorbis_I_spec.pdf
type IdentificationHeader struct {
	VorbisVersion   uint32
	AudioChannels   uint8
	AudioSampleRate uint32
	BitrateMax      uint32
	BitrateNormal   uint32
	BitrateMin      uint32
	Blocksize0      uint16
	Blocksize1      uint16
}

func ParseIdentificationHeader(b [30]byte) (IdentificationHeader, error) {
	if b[0] != 1 {
		return IdentificationHeader{}, fmt.Errorf("vorbis: invalid identification header type: %d", b[0])
	}
	if string(b[1:7]) != "vorbis" {
		return IdentificationHeader{}, fmt.Errorf("vorbis: invalid header: %s", b[1:7])
	}

	var h IdentificationHeader
	h.VorbisVersion = binary.LittleEndian.Uint32(b[7:11])
	h.AudioChannels = b[11]
	h.AudioSampleRate = binary.LittleEndian.Uint32(b[12:16])
	h.BitrateMax = binary.LittleEndian.Uint32(b[16:20])
	h.BitrateNormal = binary.LittleEndian.Uint32(b[20:24])
	h.BitrateMin = binary.LittleEndian.Uint32(b[24:28])
	h.Blocksize0 = 1 << (b[28] & 0x0F)
	h.Blocksize1 = 1 << (b[28] >> 4)
	if b[29] != 1 {
		return IdentificationHeader{}, fmt.Errorf("vorbis: invalid header framing: %s", b[29])
	}
	return h, nil
}

// IdentificationHeader is based on Section 4.2.2 of Vorbis I specification.
// See: https://xiph.org/vorbis/doc/Vorbis_I_spec.pdf
type CommentHeader struct {
	Vendor       string
	UserComments []string
}

func ParseCommentHeader(b []byte) (CommentHeader, error) {
	if b[0] != 3 {
		return CommentHeader{}, fmt.Errorf("vorbis: invalid identification header type: %d", b[0])
	}
	if string(b[1:7]) != "vorbis" {
		return CommentHeader{}, fmt.Errorf("vorbis: invalid header: %s", b[1:7])
	}

	var h CommentHeader
	b = b[7:]

	n := binary.LittleEndian.Uint32(b[:4])
	b = b[4:]
	h.Vendor = string(b[:n])
	b = b[n:]

	n = binary.LittleEndian.Uint32(b[:4])
	b = b[4:]
	h.UserComments = make([]string, n)
	for i := uint32(0); i < n; i++ {
		m := int64(binary.LittleEndian.Uint32(b[:4]))
		b = b[4:]
		h.UserComments[i] = string(b[:m])
		b = b[m:]
	}

	if b[0] != 1 {
		return CommentHeader{}, fmt.Errorf("vorbis: invalid header framing: %s", b[0])
	}

	return h, nil
}

func NewSerial() int32 {
	return rand.Int32()
}

const (
	StateInit uint8 = iota
	StateHeader
	StatePages
	StateEnd
)

// Writer implements a quirky way of writing an Ogg Vorbis document.
type Writer struct {
	w io.Writer
	b ogg.Builder

	MinSize uint32

	state  uint8
	seq    uint32
	serial int32

	totalLen uint32
	queue    [][]byte
}

func NewWriter(w io.Writer, serial int32) *Writer {
	return &Writer{
		w: w, b: ogg.NewBuilder(),
		MinSize: 4096,
		serial:  serial, state: StateInit,
	}
}

func (w *Writer) nextSeq() uint32 {
	s := w.seq
	w.seq++
	return s
}

// WriteIdentHeader initialises the logical bitstream. The writing
// has to start with this method, everything else will cause panic.
func (w *Writer) WriteIdentHeader(b []byte) error {
	// The first Vorbis packet (the identification header), which uniquely identifies a stream
	// as Vorbis audio, is placed alone in the first page of the logical Ogg stream. This
	// results in a first Ogg page of exactly 58 bytes at the very beginning of the logical
	// stream.
	if w.state != StateInit {
		panic("vorbis: invalid state")
	}
	if len(b) != 30 {
		return fmt.Errorf("matroska: Vorbis identification heade requres 30 bytes, got %d", len(b))
	}
	_, err := w.b.B().
		// This first page is marked ’beginning of stream’ in the page flags.
		HeaderType(ogg.HeaderTypeFirstPage).
		// The granule position of these first pages containing only headers is zero.
		GranulePosition(0).
		SerialNum(w.serial).
		PageSequence(w.nextSeq()).
		Segments([][]byte{b[:]}).
		WriteTo(w.w)
	if err != nil {
		return err
	}
	w.state = StateHeader
	return nil
}

// WriteHeaders directly adds the Comment and Setup headers in a single Ogg page.
//
// This method must be called after WriteIdentHeader and before any Segment call.
func (w *Writer) WriteHeaders(commentHeader, setupHeader []byte) error {
	// The second and third vorbis packets (comment and setup headers) may span one or
	// more pages beginning on the second page of the logical stream. However many pages
	// they span, the third header packet finishes the page on which it ends. The next (first
	// audio) packet must begin on a fresh page.
	if w.state != StateHeader {
		panic("vorbis: invalid state")
	}
	_, err := w.b.B().
		HeaderType(0).
		// The granule position of these first pages containing only headers is zero.
		GranulePosition(0).
		SerialNum(w.serial).
		PageSequence(w.nextSeq()).
		Segments([][]byte{commentHeader, setupHeader}).
		WriteTo(w.w)
	if err != nil {
		return err
	}
	w.state = StatePages
	return nil
}

// Segment queues segments to be written as an Ogg page. The queue is flushed
// when MinSize byte is added to the queue and when the last element is added.
//
// This function must be called after WriteHeaders.
func (w *Writer) Segment(segment []byte, granpos uint64, last bool) error {
	if w.state != StatePages {
		panic("vorbis: invalid state")
	}
	w.totalLen += uint32(len(segment))
	w.queue = append(w.queue, segment)
	if last {
		w.state = StateEnd
		if err := w.flushPage(ogg.HeaderTypeLastPage, granpos); err != nil {
			return err
		}
	}
	if w.totalLen >= w.MinSize {
		if err := w.flushPage(0, granpos); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) flushPage(headerType byte, granpos uint64) error {
	_, err := w.b.B().
		HeaderType(headerType).
		GranulePosition(granpos).
		SerialNum(w.serial).
		PageSequence(w.nextSeq()).
		Segments(w.queue).
		WriteTo(w.w)
	w.queue = w.queue[:0]
	w.totalLen = 0
	return err
}
