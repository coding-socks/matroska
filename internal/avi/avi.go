// Package avi implements the AVI file format.
//
// See: https://learn.microsoft.com/en-us/windows/win32/directshow/avi-riff-file-reference
package avi

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/coding-socks/matroska/internal/riff"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime/debug"
)

var (
	AVI = riff.FourCC{'A', 'V', 'I', ' '}

	ListINFO = riff.FourCC{'I', 'N', 'F', 'O'}
	ListHRDL = riff.FourCC{'h', 'd', 'r', 'l'}
	ListSTRL = riff.FourCC{'s', 't', 'r', 'l'}
	ListMOVI = riff.FourCC{'m', 'o', 'v', 'i'}

	ChunkAVIH = riff.FourCC{'a', 'v', 'i', 'h'} // AVI header

	ChunkSTRH = riff.FourCC{'s', 't', 'r', 'h'} // Stream header
	ChunkSTRF = riff.FourCC{'s', 't', 'r', 'f'} // Stream format
	ChunkSTRD = riff.FourCC{'s', 't', 'r', 'd'} // Stream-header data
	ChunkSTRN = riff.FourCC{'s', 't', 'r', 'n'} // Stream name
	ChunkIDX1 = riff.FourCC{'i', 'd', 'x', '1'}
	ChunkINDX = riff.FourCC{'i', 'n', 'd', 'x'}
)

var (
	StreamTypeDB = StreamType{'d', 'b'} // Uncompressed video frame
	StreamTypeDC = StreamType{'d', 'c'} // Compressed video frame
	StreamTypePC = StreamType{'p', 'c'} // Palette change
	StreamTypeWB = StreamType{'w', 'b'} // Audio data
)

type StreamType [2]byte

func NewStreamID(id uint8, streamType StreamType) riff.FourCC {
	s := small(id)
	return riff.FourCC{s[0], s[1], streamType[0], streamType[1]}
}

// small returns the string for an i with 0 <= i < 99.
func small(i uint8) string {
	if i > 99 {
		panic("matroska: invalid stream id")
	}
	return smallsString[i*2 : i*2+2]
}

const smallsString = "00010203040506070809" +
	"10111213141516171819" +
	"20212223242526272829" +
	"30313233343536373839" +
	"40414243444546474849" +
	"50515253545556575859" +
	"60616263646566676869" +
	"70717273747576777879" +
	"80818283848586878889" +
	"90919293949596979899"

type Writer struct {
	avi    *riff.Writer
	movi   *riff.Writer
	header *riff.Writer // placeholder
	idx    []byte
	err    error

	offset uint32
	maxLen uint32
}

func NewWriter(w io.WriterAt) (*Writer, error) {
	lw, err := riff.NewWriter(w, AVI)
	if err != nil {
		return nil, err
	}

	return &Writer{avi: lw}, nil
}

func (w *Writer) MaxLen() uint32 {
	return w.maxLen
}

// WriteHeader writes the header and metadata information into the
// first 2048 bytes of the given io.WriterAt.
//
// WriteHeader should be called last as it requires information like total frames.
func (w *Writer) WriteHeader(sh StreamHeader, sf StreamFormat) error {
	if w.header == nil {
		return errors.New("avi: header writer has not been initialised (forgot to call WriteData?)")
	}
	lw := w.header

	var mh MainHeader
	microSecPerFrame := float64(sh.Scale()) / float64(sh.Rate()) * float64(sh.Scale())
	mh.SetMicroSecPerFrame(uint32(math.Ceil(microSecPerFrame)))
	mh.SetFlags(AVIF_HASINDEX | AVIF_ISINTERLEAVED)
	mh.SetTotalFrames(sh.Length())
	mh.SetStreams(1) // number of audio streams plus one
	mh.SetWidth(sf.Width())
	mh.SetHeight(sf.Height())

	hdrlChunk, err := lw.Next(riff.LIST)
	if err != nil {
		return err
	}
	hdrl, err := riff.NewListWriter(hdrlChunk, ListHRDL)
	if err != nil {
		return err
	}
	{ // avih
		avih, err := hdrl.Next(ChunkAVIH)
		if err != nil {
			return err
		}
		if _, err = avih.Write(mh[:]); err != nil {
			return err
		}
	}
	{ // LIST strl
		strlChunk, err := hdrl.Next(riff.LIST)
		if err != nil {
			return err
		}
		strl, err := riff.NewListWriter(strlChunk, ListSTRL)
		if err != nil {
			return err
		}
		{ // strh
			strhChunk, err := strl.Next(ChunkSTRH)
			if err != nil {
				return err
			}
			if _, err := strhChunk.Write(sh[:]); err != nil {
				return err
			}
		}
		{ // strf
			strfChunk, err := strl.Next(ChunkSTRF)
			if err != nil {
				return err
			}
			if _, err := strfChunk.Write(sf[:]); err != nil {
				return err
			}
		}
		if err := strl.Close(); err != nil {
			return err
		}
	}
	if err := hdrl.Close(); err != nil {
		return err
	}

	infoChunk, err := lw.Next(riff.LIST)
	if err != nil {
		w.err = err
		return err
	}
	info, err := riff.NewListWriter(infoChunk, ListINFO)
	if err != nil {
		w.err = err
		return err
	}
	{
		istf, err := info.Next(riff.FourCC{'I', 'S', 'F', 'T'})
		if err != nil {
			w.err = err
			return err
		}
		buildInfo, _ := debug.ReadBuildInfo()
		if _, w.err = fmt.Fprintf(istf, "%s %s\x00", filepath.Base(os.Args[0]), buildInfo.Main.Version); w.err != nil {
			return w.err
		}
		if w.err = info.Close(); w.err != nil {
			return w.err
		}
	}
	if w.err = lw.Close(); w.err != nil {
		return w.err
	}
	return nil
}

func (w *Writer) init() error {
	if w.movi != nil {
		return errors.New("avi: movi already initialized (called WriteHeader twice?)")
	}
	lw := w.avi

	w.offset = 32 // According to my math, this is supposed to be 12
	var target uint32 = 2048
	if w.header, w.err = lw.Placeholder(target - w.offset); w.err != nil {
		return w.err
	}
	w.offset = target

	moviChunk, err := lw.Next(riff.LIST)
	if err != nil {
		return err
	}
	w.movi, err = riff.NewListWriter(moviChunk, ListMOVI)
	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) WriteData(id riff.FourCC, b []byte, flags uint32) error {
	if w.movi == nil {
		if err := w.init(); err != nil {
			return err
		}
	}
	ww, err := w.movi.Next(id)
	if err != nil {
		return err
	}
	_, err = ww.Write(b)
	if err != nil {
		return err
	}

	w.idx = append(w.idx, id[0], id[1], id[2], id[3])
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], flags)
	w.idx = append(w.idx, buf[0], buf[1], buf[2], buf[3]) // flags
	binary.LittleEndian.PutUint32(buf[:], w.offset)
	w.idx = append(w.idx, buf[0], buf[1], buf[2], buf[3]) // offset
	binary.LittleEndian.PutUint32(buf[:], uint32(len(b)))
	w.idx = append(w.idx, buf[0], buf[1], buf[2], buf[3]) // size

	w.offset += uint32(len(b)) + 8 /* id and size */
	if w.offset&1 == 1 {
		w.offset++ // padding
	}
	if l := uint32(len(b)); l > w.maxLen {
		w.maxLen = l
	}

	return nil
}

func (w *Writer) Close() error {
	if w.err != nil {
		return w.err
	}
	if w.movi != nil {
		if w.err = w.movi.Close(); w.err != nil {
			return w.err
		}
		ww, err := w.avi.Next(ChunkIDX1)
		if err != nil {
			w.err = err
			return err
		}
		if _, w.err = ww.Write(w.idx); w.err != nil {
			return w.err
		}
	}
	if w.err = w.avi.Close(); w.err != nil {
		return w.err
	}

	return w.avi.Close()
}

// MainHeader defines global information in an AVI file.
//
// See: https://learn.microsoft.com/en-us/previous-versions/windows/desktop/api/Aviriff/ns-aviriff-avimainheader
type MainHeader [56]byte

// MicroSecPerFrame specifies the number of microseconds between frames.
func (h *MainHeader) MicroSecPerFrame() uint32 {
	return binary.LittleEndian.Uint32(h[0:4])
}

func (h *MainHeader) SetMicroSecPerFrame(v uint32) {
	binary.LittleEndian.PutUint32(h[0:4], v)
}

// MaxBytesPerSec specifies the approximate maximum data rate of the file.
func (h *MainHeader) MaxBytesPerSec() uint32 {
	return binary.LittleEndian.Uint32(h[4:8])
}

func (h *MainHeader) SetMaxBytesPerSec(v uint32) {
	binary.LittleEndian.PutUint32(h[4:8], v)
}

// PaddingGranularity specifies the alignment for data, in bytes.
func (h *MainHeader) PaddingGranularity() uint32 {
	return binary.LittleEndian.Uint32(h[8:12])
}

func (h *MainHeader) SetPaddingGranularity(v uint32) {
	binary.LittleEndian.PutUint32(h[8:12], v)
}

const (
	// AVIF_HASINDEX indicates the AVI file has an index.
	AVIF_HASINDEX uint32 = 1 << 4
	// AVIF_MUSTUSEINDEX indicates that application should use the index, rather than the physical ordering of the chunks in the file, to determine the order of presentation of the data.
	AVIF_MUSTUSEINDEX uint32 = 1 << 5
	// AVIF_ISINTERLEAVED indicates the AVI file is interleaved.
	AVIF_ISINTERLEAVED uint32 = 1 << 8
	// AVIF_TRUSTCKTYPE use CKType to find key frames
	AVIF_TRUSTCKTYPE uint32 = 1 << 11
	// AVIF_WASCAPTUREFILE indicates the AVI file is a specially allocated file used for capturing real-time video.
	AVIF_WASCAPTUREFILE uint32 = 1 << 16
	// AVIF_COPYRIGHTED indicates the AVI file contains copyrighted data and software.
	AVIF_COPYRIGHTED uint32 = 1 << 17
)

func (h *MainHeader) Flags() uint32 {
	return binary.LittleEndian.Uint32(h[12:16])
}

func (h *MainHeader) SetFlags(v uint32) {
	binary.LittleEndian.PutUint32(h[12:16], v)
}

// TotalFrames specifies the total number of frames of data in the file.
func (h *MainHeader) TotalFrames() uint32 {
	return binary.LittleEndian.Uint32(h[16:20])
}

func (h *MainHeader) SetTotalFrames(v uint32) {
	binary.LittleEndian.PutUint32(h[16:20], v)
}

// InitialFrames specifies the initial frame for interleaved files.
func (h *MainHeader) InitialFrames() uint32 {
	return binary.LittleEndian.Uint32(h[20:24])
}

func (h *MainHeader) SetInitialFrames(v uint32) {
	binary.LittleEndian.PutUint32(h[20:24], v)
}

// Streams specifies the number of streams in the file.
func (h *MainHeader) Streams() uint32 {
	return binary.LittleEndian.Uint32(h[24:28])
}

func (h *MainHeader) SetStreams(v uint32) {
	binary.LittleEndian.PutUint32(h[24:28], v)
}

// SuggestedBufferSize specifies the suggested buffer size for reading the file.
func (h *MainHeader) SuggestedBufferSize() uint32 {
	return binary.LittleEndian.Uint32(h[28:32])
}

func (h *MainHeader) SetSuggestedBufferSize(v uint32) {
	binary.LittleEndian.PutUint32(h[28:32], v)
}

// Width specifies the width of the AVI file in pixels.
func (h *MainHeader) Width() uint32 {
	return binary.LittleEndian.Uint32(h[32:36])
}

func (h *MainHeader) SetWidth(v uint32) {
	binary.LittleEndian.PutUint32(h[32:36], v)
}

// Height specifies the height of the AVI file in pixels.
func (h *MainHeader) Height() uint32 {
	return binary.LittleEndian.Uint32(h[36:40])
}

func (h *MainHeader) SetHeight(v uint32) {
	binary.LittleEndian.PutUint32(h[36:40], v)
}

type StreamHeader [56]byte

var (
	StreamTypeAUDS = riff.FourCC{'a', 'u', 'd', 's'}
	StreamTypeMIDS = riff.FourCC{'m', 'i', 'd', 's'}
	StreamTypeTXTS = riff.FourCC{'t', 'x', 't', 's'}
	StreamTypeVIDS = riff.FourCC{'v', 'i', 'd', 's'}
)

// Type returns the type of the data contained in the stream.
func (h *StreamHeader) Type() riff.FourCC {
	return riff.FourCC(h[0:4])
}

func (h *StreamHeader) SetType(v riff.FourCC) {
	h[0], h[1], h[2], h[3] = v[0], v[1], v[2], v[3]
}

// Handler returns a FOURCC that optionally identifies a specific data handler.
func (h *StreamHeader) Handler() riff.FourCC {
	return riff.FourCC(h[4:8])
}

func (h *StreamHeader) SetHandler(v riff.FourCC) {
	h[4], h[5], h[6], h[7] = v[0], v[1], v[2], v[3]
}

const (
	// AVISF_DISABLED indicates this stream should not be enabled by default.
	AVISF_DISABLED uint32 = 1 << 0
	// AVISF_VIDEO_PALCHANGES indicates this video stream contains palette changes.
	AVISF_VIDEO_PALCHANGES uint32 = 1 << 16
)

func (h *StreamHeader) Flags() uint32 {
	return binary.LittleEndian.Uint32(h[8:12])
}

func (h *StreamHeader) SetFlags(v uint32) {
	binary.LittleEndian.PutUint32(h[8:12], v)
}

func (h *StreamHeader) Priority() uint16 {
	return binary.LittleEndian.Uint16(h[12:14])
}

func (h *StreamHeader) SetPriority(v uint16) {
	binary.LittleEndian.PutUint16(h[12:14], v)
}

// TODO: is this a country code?
func (h *StreamHeader) language() uint16 {
	return binary.LittleEndian.Uint16(h[14:16])
}

func (h *StreamHeader) setLanguage(v uint16) {
	binary.LittleEndian.PutUint16(h[14:16], v)
}

func (h *StreamHeader) InitialFrames() uint32 {
	return binary.LittleEndian.Uint32(h[16:20])
}

func (h *StreamHeader) SetInitialFrames(v uint32) {
	binary.LittleEndian.PutUint32(h[16:20], v)
}

func (h *StreamHeader) Scale() uint32 {
	return binary.LittleEndian.Uint32(h[20:24])
}

func (h *StreamHeader) SetScale(v uint32) {
	binary.LittleEndian.PutUint32(h[20:24], v)
}

func (h *StreamHeader) Rate() uint32 {
	return binary.LittleEndian.Uint32(h[24:28])
}

func (h *StreamHeader) SetRate(v uint32) {
	binary.LittleEndian.PutUint32(h[24:28], v)
}

func (h *StreamHeader) Start() uint32 {
	return binary.LittleEndian.Uint32(h[28:32])
}

func (h *StreamHeader) SetStart(v uint32) {
	binary.LittleEndian.PutUint32(h[28:32], v)
}

func (h *StreamHeader) Length() uint32 {
	return binary.LittleEndian.Uint32(h[32:36])
}

func (h *StreamHeader) SetLength(v uint32) {
	binary.LittleEndian.PutUint32(h[32:36], v)
}

func (h *StreamHeader) SuggestedBufferSize() uint32 {
	return binary.LittleEndian.Uint32(h[36:44])
}

func (h *StreamHeader) SetSuggestedBufferSize(v uint32) {
	binary.LittleEndian.PutUint32(h[36:44], v)
}

func (h *StreamHeader) Quality() uint32 {
	return binary.LittleEndian.Uint32(h[40:44])
}

func (h *StreamHeader) SetQuality(v uint32) {
	binary.LittleEndian.PutUint32(h[40:44], v)
}

func (h *StreamHeader) SampleSize() uint32 {
	return binary.LittleEndian.Uint32(h[44:48])
}

func (h *StreamHeader) SetSampleSize(v uint32) {
	binary.LittleEndian.PutUint32(h[44:48], v)
}

// TODO: find out the type of this.
func (h *StreamHeader) frame() uint32 {
	return binary.LittleEndian.Uint32(h[48:56])
}

func (h *StreamHeader) setFrame(v uint32) {
	binary.LittleEndian.PutUint32(h[48:56], v)
}

type StreamFormat [40]byte

func (f *StreamFormat) Size() uint32 {
	return binary.LittleEndian.Uint32(f[0:4])
}

func (f *StreamFormat) SetSize(v uint32) {
	binary.LittleEndian.PutUint32(f[0:4], v)
}

func (f *StreamFormat) Width() uint32 {
	return binary.LittleEndian.Uint32(f[4:8])
}

func (f *StreamFormat) SetWidth(v uint32) {
	binary.LittleEndian.PutUint32(f[4:8], v)
}

func (f *StreamFormat) Height() uint32 {
	return binary.LittleEndian.Uint32(f[8:12])
}

func (f *StreamFormat) SetHeight(v uint32) {
	binary.LittleEndian.PutUint32(f[8:12], v)
}

func (f *StreamFormat) Planes() uint16 {
	return binary.LittleEndian.Uint16(f[12:14])
}

func (f *StreamFormat) SetPlanes(v uint16) {
	binary.LittleEndian.PutUint16(f[12:14], v)
}

func (f *StreamFormat) BitCount() uint16 {
	return binary.LittleEndian.Uint16(f[14:16])
}

func (f *StreamFormat) SetBitCount(v uint16) {
	binary.LittleEndian.PutUint16(f[14:16], v)
}

func (f *StreamFormat) Compression() uint32 {
	return binary.LittleEndian.Uint32(f[16:20])
}

func (f *StreamFormat) SetCompression(v uint32) {
	binary.LittleEndian.PutUint32(f[16:20], v)
}

func (f *StreamFormat) SizeImage() uint32 {
	return binary.LittleEndian.Uint32(f[20:24])
}

func (f *StreamFormat) SetSizeImage(v uint32) {
	binary.LittleEndian.PutUint32(f[20:24], v)
}

func (f *StreamFormat) XPelsPerMeter() uint32 {
	return binary.LittleEndian.Uint32(f[24:28])
}

func (f *StreamFormat) SetXPelsPerMeter(v uint32) {
	binary.LittleEndian.PutUint32(f[24:28], v)
}

func (f *StreamFormat) YPelsPerMeter() uint32 {
	return binary.LittleEndian.Uint32(f[28:32])
}

func (f *StreamFormat) SetYPelsPerMeter(v uint32) {
	binary.LittleEndian.PutUint32(f[28:32], v)
}

func (f *StreamFormat) ClrUsed() uint32 {
	return binary.LittleEndian.Uint32(f[32:36])
}

func (f *StreamFormat) SetClrUsed(v uint32) {
	binary.LittleEndian.PutUint32(f[32:36], v)
}

func (f *StreamFormat) ClrImportant() uint32 {
	return binary.LittleEndian.Uint32(f[36:40])
}

func (f *StreamFormat) SetClrImportant(v uint32) {
	binary.LittleEndian.PutUint32(f[36:40], v)
}

const (
	// AVIIF_LIST indicates that the data chunk is a 'rec ' list.
	AVIIF_LIST uint32 = 1 << 0
	// AVIIF_KEYFRAME indicates that the data chunk is a key frame.
	AVIIF_KEYFRAME uint32 = 1 << 4
	// AVIIF_NO_TIME indicates that the data chunk does not affect the timing of the stream.
	AVIIF_NO_TIME uint32 = 1 << 8
)
