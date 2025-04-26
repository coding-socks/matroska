// Package riff implements the RIFF bitstream format.
//
// A RIFF bitstream contains a sequence of chunks. Each chunk consists of an 8-byte
// header (containing a 4-byte chunk type and a 4-byte chunk length), the chunk
// data (presented as an io.Reader), and some padding bytes.
//
// See: http://www.tactilemedia.com/info/MCI_Control_Info.html
package riff // import "golang.org/x/image/riff"

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
)

var (
	ErrMissingRIFFMagicNumber = errors.New("riff: missing RIFF magic number")
	ErrDataTooLong            = errors.New("riff: data requires more than 2^32 bytes")

	errMissingPaddingByte  = errors.New("riff: missing padding byte")
	errListSubchunkTooLong = errors.New("riff: list subchunk too long")
	errShortChunk          = errors.New("riff: short chunk")
	errStaleReader         = errors.New("riff: stale reader")
	errStaleWriter         = errors.New("riff: stale writer")
	errInvalidOffset       = errors.New("riff: invalid offset")
)

// FourCC is a four character code.
type FourCC [4]byte

// LIST is the "LIST" FourCC.
var LIST = FourCC{'L', 'I', 'S', 'T'}

// NewReader returns the initial RIFF list reader as a *Reader.
func NewReader(r io.Reader) (FourCC, *Reader, error) {
	var buf [8]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			err = ErrMissingRIFFMagicNumber
		}
		return FourCC{}, nil, err
	}
	if buf[0] != 'R' || buf[1] != 'I' || buf[2] != 'F' || buf[3] != 'F' {
		return FourCC{}, nil, ErrMissingRIFFMagicNumber
	}
	return NewListReader(binary.LittleEndian.Uint32(buf[4:]), r)
}

// NewListReader returns a LIST list reader as a *Reader.
func NewListReader(chunkLen uint32, chunkData io.Reader) (FourCC, *Reader, error) {
	if chunkLen < 4 {
		return FourCC{}, nil, errShortChunk
	}
	r := &Reader{r: &io.LimitedReader{R: chunkData, N: int64(chunkLen)}, len: chunkLen}
	if _, err := io.ReadFull(r.r, r.buf[:4]); err != nil {
		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			err = errors.Join(io.ErrUnexpectedEOF, errShortChunk)
		}
		return FourCC{}, nil, err
	}
	return FourCC{r.buf[0], r.buf[1], r.buf[2], r.buf[3]}, r, nil
}

// Reader reads chunks from an underlying io.Reader.
type Reader struct {
	r   *io.LimitedReader
	err error

	buf [8]byte

	len    uint32
	padded bool
	chunk  *ChunkReader
}

type ChunkReader struct {
	listr *Reader
	r     *io.LimitedReader
}

func (c *ChunkReader) Read(p []byte) (n int, err error) {
	if c.listr.chunk != c {
		return 0, errStaleReader
	}
	return c.r.Read(p)
}

// Next returns the next chunk. The io.Reader of the element returned becomes
// stale after the next Next call, and should no longer be used.
//
// When Next encounters io.EOF or end of the chunk, it returns io.EOF.
func (r *Reader) Next() (FourCC, uint32, *ChunkReader, error) {
	if r.err != nil {
		return FourCC{}, 0, nil, r.err
	}

	r.chunk = nil
	// If the chunk size is an odd number of bytes, a pad byte
	// with value zero is written after ckData.
	if r.padded {
		if _, r.err = io.ReadFull(r.r, r.buf[:1]); r.err != nil {
			if r.err == io.EOF {
				r.err = errors.Join(io.ErrUnexpectedEOF, errMissingPaddingByte)
			}
			return FourCC{}, 0, nil, r.err
		}
	}

	if n, err := io.ReadFull(r.r, r.buf[:]); err != nil {
		if (n != 0 && err == io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			err = errors.Join(io.ErrUnexpectedEOF, errShortChunk)
		}
		r.err = err
		return FourCC{}, 0, nil, err
	}
	id := FourCC{r.buf[0], r.buf[1], r.buf[2], r.buf[3]}
	l := binary.LittleEndian.Uint32(r.buf[4:])
	if l > r.len {
		r.err = errListSubchunkTooLong
		return FourCC{}, 0, nil, r.err
	}
	r.padded = l&1 == 1
	cr := ChunkReader{listr: r, r: &io.LimitedReader{R: r.r, N: int64(l)}}
	r.chunk = &cr
	return id, l, &cr, nil
}

func NewWriter(w io.WriterAt, fileType FourCC) (data *Writer, err error) {
	if _, err := w.WriteAt([]byte{'R', 'I', 'F', 'F'}, 0); err != nil {
		return nil, err
	}
	ww, err := NewListWriter(io.NewOffsetWriter(w, 8), fileType)
	if err != nil {
		return ww, err
	}
	ww.root = w
	return ww, err
}

func NewListWriter(chunkData io.WriterAt, listType FourCC) (listw *Writer, err error) {
	w := &Writer{w: chunkData, len: 4}
	_, w.err = chunkData.WriteAt(listType[:], 0)
	return w, w.err
}

type Writer struct {
	w    io.WriterAt
	root io.WriterAt
	err  error

	chunk  *ChunkWriter
	closed bool
	len    int64
}

type ChunkWriter struct {
	listw *Writer
	len   int64

	base int64 // the original offset
	off  int64 // the current offset
}

func (w *ChunkWriter) Write(p []byte) (n int, err error) {
	if w.listw.chunk != w {
		return 0, errStaleWriter
	}
	n, err = w.WriteAt(p, w.off)
	w.off += int64(n)
	return
}

func (w *ChunkWriter) WriteAt(p []byte, off int64) (n int, err error) {
	if w.listw.chunk != w {
		return 0, errStaleWriter
	}
	if off < 0 {
		return 0, errInvalidOffset
	}
	if (w.base + w.len + int64(len(p))) > math.MaxUint32 {
		return 0, ErrDataTooLong
	}

	n, err = w.listw.w.WriteAt(p, off+w.base)
	l := off + int64(n)
	if l > w.len {
		w.len = l
		w.listw.len = w.base + w.len
	}
	return n, err
}

func (w *ChunkWriter) close() error {
	var sizebuf [4]byte
	l := uint32(w.len)
	binary.LittleEndian.PutUint32(sizebuf[:], l)
	if _, err := w.listw.w.WriteAt(sizebuf[:], w.base-4); err != nil {
		return err
	}
	if l&1 == 1 {
		n, err := w.listw.w.WriteAt([]byte{0}, w.base+w.len)
		w.len += int64(n)
		w.listw.len += int64(n)
		if err != nil {
			return err
		}
	}
	return nil
}

var sizePlaceholder = []byte{0, 0, 0, 0}

// Next returns a chunk writer which behaves as an io.OffsetWriter.
//
// Calling Next closes the returned writer.
func (w *Writer) Next(id FourCC) (*ChunkWriter, error) {
	if w.err != nil {
		return nil, w.err
	}

	if w.chunk != nil {
		if w.err = w.chunk.close(); w.err != nil {
			return nil, w.err
		}
		w.chunk = nil
	}

	n, err := w.w.WriteAt(id[:], w.len)
	w.len += int64(n)
	if err != nil {
		w.err = err
		return nil, w.err
	}
	n, err = w.w.WriteAt(sizePlaceholder, w.len)
	w.len += int64(n)
	if err != nil {
		w.err = err
		return nil, w.err
	}
	w.chunk = &ChunkWriter{listw: w, base: w.len}
	return w.chunk, nil
}

// Close closes the list by closing the last writer returned by Next.
func (w *Writer) Close() error {
	if w.err != nil {
		return w.err
	}
	if w.closed {
		return nil
	}

	if w.chunk != nil {
		if w.err = w.chunk.close(); w.err != nil {
			return w.err
		}
		w.chunk = nil
	}

	if w.root != nil {
		// TODO: think of a better way of writing the size of the root element,
		//   but not write it for sublists because that occurs in ChunkWriter.
		var sizebuf [4]byte
		l := uint32(w.len)
		binary.LittleEndian.PutUint32(sizebuf[:], l)
		if _, w.err = w.root.WriteAt(sizebuf[:], 4); w.err != nil {
			return w.err
		}
		if l&1 == 1 {
			n, err := w.w.WriteAt([]byte{0}, w.len)
			w.len += int64(n)
			if err != nil {
				return err
			}
		}
	}
	w.closed = true
	return nil
}
