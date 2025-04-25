// Package riff implements the RIFF bitstream format.
//
// A RIFF bitstream contains a sequence of chunks. Each chunk consists of an 8-byte
// header (containing a 4-byte chunk type and a 4-byte chunk length), the chunk
// data (presented as an io.Reader), and some padding bytes.
//
// See: http://www.tactilemedia.com/info/MCI_Control_Info.html
package riff // import "golang.org/x/image/riff"

import (
	"bytes"
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
)

var u32 = binary.LittleEndian.Uint32

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
	return NewListReader(u32(buf[4:]), r)
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
	l := u32(r.buf[4:])
	if l > r.len {
		r.err = errListSubchunkTooLong
		return FourCC{}, 0, nil, r.err
	}
	r.padded = l&1 == 1
	cr := ChunkReader{listr: r, r: &io.LimitedReader{R: r.r, N: int64(l)}}
	r.chunk = &cr
	return id, l, &cr, nil
}

func NewWriter(w io.Writer, fileType FourCC) (data *Writer, err error) {
	if _, err := w.Write([]byte{'R', 'I', 'F', 'F'}); err != nil {
		return nil, err
	}
	ww, err := NewListWriter(w, fileType)
	if err != nil {
		return ww, err
	}
	ww.root = true
	return ww, err
}

func NewListWriter(chunkData io.Writer, listType FourCC) (listw *Writer, err error) {
	w := &Writer{w: chunkData}
	_, w.err = w.buf.Write(listType[:])
	return w, w.err
}

type Writer struct {
	w   io.Writer
	err error

	chunk *ChunkWriter
	buf   bytes.Buffer
	root  bool
}

type ChunkWriter struct {
	listw *Writer
	buf   bytes.Buffer
}

func (w *ChunkWriter) Write(p []byte) (n int, err error) {
	if w.listw.chunk != w {
		return 0, errStaleWriter
	}
	if (w.buf.Len() + len(p)) > math.MaxUint32 {
		return 0, ErrDataTooLong
	}
	return w.buf.Write(p)
}

func (w *ChunkWriter) close() error {
	var sizebuf [4]byte
	l := uint32(w.buf.Len())
	binary.LittleEndian.PutUint32(sizebuf[:], l)
	if _, err := w.listw.buf.Write(sizebuf[:]); err != nil {
		return err
	}
	if l&1 == 1 {
		if err := w.buf.WriteByte(0); err != nil {
			return err
		}
	}
	_, err := w.buf.WriteTo(&w.listw.buf)
	if err != nil {
		return err
	}
	return err
}

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

	if l := w.buf.Len(); l&1 == 1 {
		if _, w.err = w.buf.Write([]byte{0}); w.err != nil {
			return nil, w.err
		}
	}
	if _, w.err = w.buf.Write(id[:]); w.err != nil {
		return nil, w.err
	}
	w.chunk = &ChunkWriter{listw: w}
	return w.chunk, nil
}

func (w *Writer) Close() error {
	if w.chunk != nil {
		if w.err = w.chunk.close(); w.err != nil {
			return w.err
		}
		w.chunk = nil
	}

	if w.root {
		// TODO: think of a better way of writing the size of the root element,
		//   but not write it for sublists because that occurs in ChunkWriter.
		var sizebuf [4]byte
		l := uint32(w.buf.Len())
		binary.LittleEndian.PutUint32(sizebuf[:], l)
		if _, err := w.w.Write(sizebuf[:]); err != nil {
			return err
		}
		if l&1 == 1 {
			if err := w.buf.WriteByte(0); err != nil {
				return err
			}
		}
	}
	_, err := w.buf.WriteTo(w.w)
	if err != nil {
		return err
	}
	return err
}
