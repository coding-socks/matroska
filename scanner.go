package matroska

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/coding-socks/ebml"
	"io"
)

// ErrUnexpectedClusterElement means that Cluster was encountered before
// an Info and a Tracks Element or a SeekHead Element.
//
// The first Info Element and the first Tracks Element MUST either be
// stored before the first Cluster Element or both SHALL be referenced
// by a SeekHead Element occurring before the first Cluster Element.
var ErrUnexpectedClusterElement = errors.New("unexpected Cluster")

type Scanner struct {
	decoder *ebml.Decoder
	header  *ebml.EBML

	segmentEl    ebml.Element
	segmentStart int64

	info     *Info
	tracks   *Tracks
	seekHead *SeekHead
	// fSeekHead is an attempt to recreate SeekHead in case it is missing.
	fSeekHead *SeekHead

	offset  int64
	cluster Cluster
	err     error
}

func NewScanner(r io.ReadSeeker) (*Scanner, error) {
	d := ebml.NewDecoder(r)
	h, err := d.DecodeHeader()
	if err != nil {
		return nil, fmt.Errorf("matroska: could not decode header: %w", err)
	}
	if h.DocType != DocType {
		return nil, fmt.Errorf("matroska: cannot decode DocType: %v", h.DocType)
	}
	def, err := ebml.Definition(h.DocType)
	if err != nil {
		panic("matroska: document type is not registered")
	}
	s := Scanner{
		decoder: d,
		header:  h,

		fSeekHead: &SeekHead{},
	}
	if err := s.init(def); err != nil {
		return nil, err
	}
	return &s, nil
}

// Header returns the ebml.EBML element of the matroska document
func (s *Scanner) Header() *ebml.EBML {
	return s.header
}

// Info returns the Info element of the matroska document.
func (s *Scanner) Info() *Info {
	return s.info
}

// Tracks returns the Tracks element of the matroska document.
func (s *Scanner) Tracks() *Tracks {
	return s.tracks
}

// SeekHead returns the SeekHead element of the matroska document
// or a constructed version of it. The second return value is false
// when it cannot be trusted because it is constructed.
func (s *Scanner) SeekHead() (*SeekHead, bool) {
	if s.seekHead == nil {
		return s.fSeekHead, false
	}
	return s.seekHead, true
}

func (s *Scanner) Next() bool {
	d := s.decoder
	segmentEl := s.segmentEl
	for {
		el, n, err := d.NextOf(segmentEl, s.offset)
		if segmentEl.DataSize.Known() {
			s.offset += int64(n)
		}
		if err == ebml.ErrInvalidVINTLength {
			continue
		} else if err == io.EOF {
			return false
		} else if err != nil {
			s.err = err
			return false
		}
		if segmentEl.DataSize.Known() {
			// detect element overflow early to pretend the element is smaller
			if segmentEl.DataSize.Size() < s.offset+el.DataSize.Size() {
				el.DataSize = ebml.NewKnownDataSize(segmentEl.DataSize.Size() - s.offset)
			}
			s.offset += el.DataSize.Size()
		}
		switch el.ID {
		default:
			if _, err := d.Seek(el.DataSize.Size(), io.SeekCurrent); err != nil {
				s.err = fmt.Errorf("matroska: could not skip %s: %w", el.ID, err)
				return false
			}
		case IDCluster:
			var cl Cluster
			if err := d.Decode(&cl); err != nil {
				s.err = fmt.Errorf("matroska: could not decode %s: %s", el.ID, err)
				return err == ebml.ErrElementOverflow
			}
			s.cluster = cl
			return true
		}
	}
}

func (s *Scanner) Cluster() Cluster {
	return s.cluster
}

func (s *Scanner) Err() error {
	return s.err
}

func (s *Scanner) init(def *ebml.Def) error {
	// find root element
segment:
	for {
		el, _, err := s.decoder.Next()
		if err != nil {
			return fmt.Errorf("matroska: %w", err)
		}
		switch el.ID {
		default:
			return fmt.Errorf("matroska: got %s instead of segment", el.ID)
		case ebml.IDVoid:
			continue
		case def.Root.ID:
			s.segmentEl = el
			break segment // Done here
		}
	}
	s.segmentStart, _ = s.decoder.Seek(0, io.SeekCurrent)

	var offset int64
	for {
		el, n, err := s.decoder.NextOf(s.segmentEl, offset)
		if s.segmentEl.DataSize.Known() {
			offset += int64(n)
		}
		if err == ebml.ErrInvalidVINTLength {
			continue
		} else if err == io.EOF {
			return io.ErrUnexpectedEOF
		} else if err != nil {
			return fmt.Errorf("matroska: could not decode element: %w", err)
		}
		if s.segmentEl.DataSize.Known() {
			// detect element overflow early to pretend the element is smaller
			if s.segmentEl.DataSize.Size() < s.offset+el.DataSize.Size() {
				el.DataSize = ebml.NewKnownDataSize(s.segmentEl.DataSize.Size() - s.offset)
			}
			offset += el.DataSize.Size()
		}
		if err := s.updateFSeek(el); err != nil {
			return err
		}
		switch el.ID {
		default:
			if _, err := s.decoder.Seek(el.DataSize.Size(), io.SeekCurrent); err != nil {
				return fmt.Errorf("matroska: could not skip %s: %w", el.ID, err)
			}
			continue
		case IDSeekHead:
			sh := &SeekHead{}
			if err := s.decoder.Decode(sh); err != nil {
				return fmt.Errorf("matroska: could not decode %s: %s", el.ID, err)
			}
			// There could be a second SeekHead element according to Section 6.3.
			if s.seekHead == nil {
				s.seekHead = sh
				if o, found := s.seekTo(IDSeekHead, 0); found {
					offset = o
					continue
				}
			} else {
				s.seekHead.Seek = append(s.seekHead.Seek, sh.Seek...)
			}
		case IDInfo:
			s.info = &Info{}
			if err := s.decoder.Decode(s.info); err != nil {
				return fmt.Errorf("matroska: could not decode %s: %s", el.ID, err)
			}
		case IDTracks:
			s.tracks = &Tracks{}
			if err := s.decoder.Decode(s.tracks); err != nil {
				return fmt.Errorf("matroska: could not decode %s: %s", el.ID, err)
			}
		case IDCluster:
			return ErrUnexpectedClusterElement
		}

		if s.seekHead != nil && s.info == nil {
			offset, _ = s.seekTo(IDInfo, 0)
			continue
		}
		if s.seekHead != nil && s.tracks == nil {
			offset, _ = s.seekTo(IDTracks, 0)
			continue
		}
		if s.info != nil && s.tracks != nil {
			break
		}
	}
	if s.seekHead != nil {
		offset, _ = s.seekTo(IDCluster, 0)
	} else {
		// find cluster element
	cluster:
		for {
			el, n, err := s.decoder.Next()
			if s.segmentEl.DataSize.Known() {
				offset += int64(n)
			}
			if err != nil {
				return fmt.Errorf("matroska: %w", err)
			}
			if err := s.updateFSeek(el); err != nil {
				return err
			}
			if s.segmentEl.DataSize.Known() {
				// detect element overflow early to pretend the element is smaller
				if s.segmentEl.DataSize.Size() < s.offset+el.DataSize.Size() {
					el.DataSize = ebml.NewKnownDataSize(s.segmentEl.DataSize.Size() - s.offset)
				}
				offset += el.DataSize.Size()
			}
			switch el.ID {
			default:
				if _, err := s.decoder.Seek(el.DataSize.Size(), io.SeekCurrent); err != nil {
					return fmt.Errorf("matroska: could not skip %s: %w", el.ID, err)
				}
				continue
			case IDCluster:
				if _, err := s.decoder.Seek(int64(-n), io.SeekCurrent); err != nil {
					return fmt.Errorf("matroska: could not revert read: %w", err)
				}
				break cluster // Done here
			}
		}
	}
	s.offset = offset
	return nil
}

func (s *Scanner) seekTo(seekID string, n int) (int64, bool) {
	if s.seekHead == nil {
		return 0, false
	}
	i := 0
	for _, seek := range s.seekHead.Seek {
		id := fmt.Sprintf("0x%X", seek.SeekID)
		if id == seekID {
			if i < n {
				i++
				continue
			}
			s.decoder.Seek(s.segmentStart+int64(seek.SeekPosition), io.SeekStart)
			return int64(seek.SeekPosition), true
		}
	}
	return 0, false
}

func (s *Scanner) updateFSeek(el ebml.Element) error {
	if s.seekHead == nil && el.ID != IDSeekHead && el.ID != ebml.IDVoid && el.ID != ebml.IDCRC32 {
		pos, err := s.decoder.Seek(0, io.SeekCurrent)
		if err != nil {
			return fmt.Errorf("matroska: could not read position: %w", err)
		}
		id, _ := hex.DecodeString(el.ID[2:])
		s.fSeekHead.Seek = append(s.fSeekHead.Seek, Seek{
			SeekID:       id,
			SeekPosition: uint(pos - s.segmentStart),
		})
	}
	return nil
}
