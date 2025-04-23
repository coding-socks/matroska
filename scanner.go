package matroska

import (
	"errors"
	"fmt"
	"github.com/coding-socks/ebml"
	"github.com/coding-socks/ebml/schema"
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

	offset       int64
	cluster      Cluster
	firstCluster *Cluster
	err          error
}

func NewScanner(r io.Reader) *Scanner {
	d := ebml.NewDecoder(r)
	s := Scanner{
		decoder: d,
	}
	return &s
}

func (s *Scanner) Decoder() *ebml.Decoder {
	return s.decoder
}

func (s *Scanner) Init() error {
	if s.err != nil || s.header != nil {
		return s.err
	}
	h, err := s.decoder.DecodeHeader()
	if err != nil {
		return fmt.Errorf("matroska: could not decode header: %w", err)
	}
	s.header = h
	if h.DocType != DocType {
		return fmt.Errorf("matroska: cannot decode DocType: %v", h.DocType)
	}
	def, err := ebml.Definition(h.DocType)
	if err != nil {
		panic("matroska: document type is not registered")
	}
	s.fSeekHead = &SeekHead{}
	if err := s.init(def); err != nil {
		return err
	}
	return nil
}

// Header returns the ebml.EBML element of the matroska document
func (s *Scanner) Header() *ebml.EBML {
	s.err = s.Init()
	return s.header
}

// Info returns the Info element of the matroska document.
func (s *Scanner) Info() *Info {
	s.err = s.Init()
	return s.info
}

// Tracks returns the Tracks element of the matroska document.
func (s *Scanner) Tracks() *Tracks {
	s.err = s.Init()
	return s.tracks
}

// SeekHead returns the SeekHead element of the matroska document
// or a constructed version of it. The second return value is false
// when it cannot be trusted because it is constructed.
func (s *Scanner) SeekHead() (*SeekHead, bool) {
	s.err = s.Init()
	if s.seekHead == nil {
		return s.fSeekHead, false
	}
	return s.seekHead, true
}

// Next reads the next Cluster struct from the io.Reader.
//
// The cluster is accessible by calling Cluster.
func (s *Scanner) Next() bool {
	s.err = s.Init()
	if s.err != nil {
		return false
	}
	if s.firstCluster != nil {
		s.cluster = *s.firstCluster
		s.firstCluster = nil
		return true
	}
	return s.next()
}

func (s *Scanner) next() bool {
	d := s.decoder
	segmentEl := s.segmentEl
	var offset int64 = 0
	for {
		el, n, err := d.NextOf(segmentEl, offset)
		if segmentEl.DataSize != -1 {
			offset += int64(n)
		}
		if errors.Is(err, ebml.ErrInvalidVINTLength) {
			_ = d.SkipByte()
			offset += 1
			continue
		} else if errors.Is(err, ebml.ErrElementOverflow) {
			// detect element overflow early to pretend the element is smaller
			if segmentEl.DataSize < offset+el.DataSize {
				el.DataSize = segmentEl.DataSize - offset
			}
		} else if err == io.EOF {
			return false
		} else if err != nil {
			s.err = err
			return false
		}
		if segmentEl.DataSize != -1 {
			offset += el.DataSize
		}
		switch el.ID {
		default:
			if err := d.Skip(el); err != nil {
				s.err = fmt.Errorf("matroska: could not skip %v: %w", el.ID, err)
				return false
			}
		case IDChapters: // TODO: populate chapters
			if err := s.updateFSeek(el); err != nil && !errors.Is(err, errors.ErrUnsupported) {
				s.err = err
				return false
			}
			var chapters Chapters
			if err := s.decoder.Decode(el, &chapters); err != nil {
				s.err = fmt.Errorf("matroska: could not decode %v: %w", el.ID, err)
				return false
			}
		case IDCues: // TODO: populate cues
			if err := s.updateFSeek(el); err != nil && !errors.Is(err, errors.ErrUnsupported) {
				s.err = err
				return false
			}
			var cues Cues
			if err := s.decoder.Decode(el, &cues); err != nil {
				s.err = fmt.Errorf("matroska: could not decode %v: %w", el.ID, err)
				return false
			}
		case IDAttachments: // TODO: populate attachments
			if err := s.updateFSeek(el); err != nil && !errors.Is(err, errors.ErrUnsupported) {
				s.err = err
				return false
			}
			var attachments Attachments
			if err := s.decoder.Decode(el, &attachments); err != nil {
				s.err = fmt.Errorf("matroska: could not decode %v: %w", el.ID, err)
				return false
			}
		case IDTags: // TODO: populate tags
			if err := s.updateFSeek(el); err != nil && !errors.Is(err, errors.ErrUnsupported) {
				s.err = err
				return false
			}
			var tags Tags
			if err := s.decoder.Decode(el, &tags); err != nil {
				s.err = fmt.Errorf("matroska: could not decode %v: %w", el.ID, err)
				return false
			}

		case IDCluster:
			var cl Cluster
			if err := d.Decode(el, &cl); err != nil {
				if !errors.Is(err, ebml.ErrElementOverflow) {
					s.err = fmt.Errorf("matroska: could not decode %v: %w", el.ID, err)
					return false
				}
			}
			s.cluster = cl
			return true
		}
	}
}

// Cluster returns the latest Cluster struct read from the io.Reader.
func (s *Scanner) Cluster() Cluster {
	return s.cluster
}

// Err returns any errors detected while reading the io.Reader.
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
			return fmt.Errorf("matroska: got %v instead of segment", el.ID)
		case ebml.IDVoid:
			continue
		case def.Root.ID:
			s.segmentEl = el
			break segment // Done here
		}
	}
	if ss, ok := s.decoder.AsSeeker(); ok {
		s.segmentStart, _ = ss.Seek(0, io.SeekCurrent)
	}

	var offset int64
	for {
		el, n, err := s.decoder.NextOf(s.segmentEl, offset)
		if s.segmentEl.DataSize != -1 {
			offset += int64(n)
		}
		if errors.Is(err, ebml.ErrInvalidVINTLength) {
			_ = s.decoder.SkipByte()
			offset += 1
			continue
		} else if err == io.EOF {
			return io.ErrUnexpectedEOF
		} else if err != nil {
			return fmt.Errorf("matroska: could not decode element: %w", err)
		}
		if s.segmentEl.DataSize != -1 {
			// detect element overflow early to pretend the element is smaller
			if s.segmentEl.DataSize < s.offset+el.DataSize {
				el.DataSize = s.segmentEl.DataSize - s.offset
			}
			offset += el.DataSize
		}
		if err := s.updateFSeek(el); err != nil && !errors.Is(err, errors.ErrUnsupported) {
			return err
		}
		switch el.ID {
		default:
			if err := s.decoder.Skip(el); err != nil {
				return fmt.Errorf("matroska: could not skip %v: %w", el.ID, err)
			}
			continue
		case IDSeekHead:
			sh := &SeekHead{}
			if err := s.decoder.Decode(el, sh); err != nil {
				return fmt.Errorf("matroska: could not decode %v: %w", el.ID, err)
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
			if err := s.decoder.Decode(el, s.info); err != nil {
				return fmt.Errorf("matroska: could not decode %v: %w", el.ID, err)
			}
		case IDTracks:
			s.tracks = &Tracks{}
			if err := s.decoder.Decode(el, s.tracks); err != nil {
				return fmt.Errorf("matroska: could not decode %v: %w", el.ID, err)
			}
		case IDCluster:
			return ErrUnexpectedClusterElement
		}

		if s.seekHead != nil && s.info == nil {
			var ok bool
			if offset, ok = s.seekTo(IDInfo, 0); ok {
				continue
			}
		}
		if s.seekHead != nil && s.tracks == nil {
			var ok bool
			if offset, ok = s.seekTo(IDTracks, 0); ok {
				continue
			}
		}
		if s.info != nil && s.tracks != nil {
			break
		}
	}

	if s.seekHead != nil {
		var ok bool
		if offset, ok = s.seekTo(IDCluster, 0); ok {
			s.offset = offset
			return nil
		}
	}

	s.offset = offset
	// find cluster element
	_ = s.next()
	cluster := s.Cluster()
	s.firstCluster = &cluster
	return s.err
}

func (s *Scanner) seekTo(seekID schema.ElementID, n int) (int64, bool) {
	if s.seekHead == nil {
		return 0, false
	}
	ss, ok := s.decoder.AsSeeker()
	if !ok {
		return 0, false
	}
	i := 0
	for _, seek := range s.seekHead.Seek {
		if seek.SeekID == seekID {
			if i < n {
				i++
				continue
			}
			ss.Seek(s.segmentStart+int64(seek.SeekPosition), io.SeekStart)
			return int64(seek.SeekPosition), true
		}
	}
	return 0, false
}

func (s *Scanner) updateFSeek(el ebml.Element) error {
	if s.seekHead == nil && el.ID != IDSeekHead && el.ID != ebml.IDVoid && el.ID != ebml.IDCRC32 {
		ss, ok := s.decoder.AsSeeker()
		if !ok {
			return errors.ErrUnsupported
		}
		pos, err := ss.Seek(0, io.SeekCurrent)
		if err != nil {
			return fmt.Errorf("matroska: could not read position: %w", err)
		}
		s.fSeekHead.Seek = append(s.fSeekHead.Seek, Seek{
			SeekID:       el.ID,
			SeekPosition: uint(pos - s.segmentStart),
		})
	}
	return nil
}
