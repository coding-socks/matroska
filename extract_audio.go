package matroska

import (
	"fmt"
	"github.com/coding-socks/matroska/internal/vorbis"
	"io"
	"math/rand/v2"
)

func extractTrackAudio(w io.Writer, s *Scanner, t TrackEntry) error {
	switch t.CodecID {
	case AudioCodecMP2, AudioCodecMP3:
		return extractTrackMPEG(w, s, t)
	case AudioCodecVORBIS:
		return extractTrackVORBIS(w, s, t)
	}
	return fmt.Errorf("matroska: unknown audio codec %s", t.CodecID)
}

func extractTrackMPEG(w io.Writer, s *Scanner, t TrackEntry) error {
	for s.Next() {
		c := s.Cluster()
		m := len(c.SimpleBlock) + len(c.BlockGroup)
		if m == 0 {
			continue
		}
		for i := range c.SimpleBlock {
			block, err := ReadSimpleBlock(c.SimpleBlock[i], c.Timestamp)
			if err != nil {
				return fmt.Errorf("matroska: could not create block struct: %w", err)
			}
			if block.TrackNumber() != t.TrackNumber {
				continue
			}
			if _, err := io.Copy(w, block.Data()); err != nil {
				return err
			}
		}
		for i := range c.BlockGroup {
			block, err := ReadBlock(c.BlockGroup[i].Block, c.Timestamp)
			if err != nil {
				return fmt.Errorf("matroska: could not create block struct: %w", err)
			}
			if block.TrackNumber() != t.TrackNumber {
				continue
			}
			if _, err := io.Copy(w, block.Data()); err != nil {
				return err
			}
		}
	}
	return nil
}

// extractTrackVORBIS is based on the Vorbis I specification created by the Xiph.Org Foundation.
// See: https://xiph.org/vorbis/doc/Vorbis_I_spec.pdf
func extractTrackVORBIS(w io.Writer, s *Scanner, track TrackEntry) error {
	serialNum := rand.Int32()

	vw := vorbis.NewWriter(w, serialNum)

	codecPrivate := *track.CodecPrivate
	frames := Frames(LacingFlagXiph, codecPrivate)
	if len(frames) != 3 {
		return fmt.Errorf("matroska: Vorbis audio track requires 3 header pages, got %d", len(frames))
	}

	ih := frames[0]
	if err := vw.WriteIdentHeader(ih); err != nil {
		return err
	}
	iheader, err := vorbis.ParseIdentificationHeader([30]byte(ih))
	if err != nil {
		return err
	}
	blockSizes := []uint16{
		iheader.Blocksize0,
		iheader.Blocksize1,
	}
	cheader, err := vorbis.ParseCommentHeader(frames[1])
	if err != nil {
		return err
	}
	_ = cheader

	if err := vw.WriteHeaders(frames[1], frames[2]); err != nil {
		return err
	}

	var (
		prevBlockSize uint64 = 0
		granpos       uint64 = 0

		prevFrame []byte //
	)
	frames = frames[:0] // clear value

	for s.Next() {
		c := s.Cluster()
		if (len(c.SimpleBlock) + len(c.BlockGroup)) == 0 {
			continue
		}

		for i := range c.SimpleBlock {
			block, err := ReadSimpleBlock(c.SimpleBlock[i], c.Timestamp)
			if err != nil {
				return fmt.Errorf("matroska: could not create block struct: %w", err)
			}
			if block.TrackNumber() != track.TrackNumber {
				continue
			}
			frames = append(frames, block.Frames()...)
		}
		for i := range c.BlockGroup {
			block, err := ReadBlock(c.BlockGroup[i].Block, c.Timestamp)
			if err != nil {
				return fmt.Errorf("matroska: could not create block struct: %w", err)
			}
			if block.TrackNumber() != track.TrackNumber {
				continue
			}
			frames = append(frames, block.Frames()...)
		}

		for _, frame := range frames {
			blockSize := uint64(blockSizes[(frame[0]>>1)&1])

			if prevFrame != nil {
				if err := vw.Segment(prevFrame, granpos, false); err != nil {
					return err
				}

				// We need at least two segment to calculate this
				granpos += (blockSize + prevBlockSize) / 4
			}

			prevBlockSize = blockSize
			prevFrame = frame
		}

		frames = frames[:0]
	}
	if prevFrame == nil {
		return nil
	}
	// only this element can be the last.
	return vw.Segment(prevFrame, granpos, true)
}
