package matroska

import (
	"encoding/binary"
	"fmt"
	"github.com/coding-socks/matroska/internal/avi"
	"github.com/coding-socks/matroska/internal/riff"
	"io"
	"math"
	"os"
)

func extractTrackVideo(w *os.File, s *Scanner, t TrackEntry) error {
	switch t.CodecID {
	case VideoCodecMSCOMP:
		return extractTrackMSCOMP(w, s, t)
	}
	return fmt.Errorf("matroska: unknown audio codec %s", t.CodecID)
}

func extractTrackMSCOMP(w io.WriterAt, s *Scanner, t TrackEntry) error {
	scale := s.Info().TimestampScale
	if t.Video == nil {
		return fmt.Errorf("matroska: missing video stream")
	}

	ww, err := avi.NewWriter(w)
	if err != nil {
		return fmt.Errorf("matroska: could not to initiate avi file: %w", err)
	}
	defer ww.Close()

	videoStreamID := avi.NewStreamID(0, avi.StreamTypeDC)
	var totalFrames uint32 = 0
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
			frames := block.Frames()
			for i := range frames {
				var flags uint32 = 0
				if i == 0 && ((block.Flags() & SimpleBlockFlagKeyframe) > 0) {
					flags |= avi.AVIIF_KEYFRAME
				}
				if err := ww.WriteData(videoStreamID, frames[i], flags); err != nil {
					return err
				}
				totalFrames++
			}
		}
		for _, group := range c.BlockGroup {
			block, err := ReadBlock(group.Block, c.Timestamp)
			if err != nil {
				return fmt.Errorf("matroska: could not create block struct: %w", err)
			}
			if block.TrackNumber() != t.TrackNumber {
				continue
			}
			frames := block.Frames()
			for i := range frames {
				var flags uint32 = 0
				// TODO: I'm not sure if this is correct. Maybe this is only relevant for
				//  RAPs (i.e., frames that don't depend on other frames).
				if i == 0 && len(group.ReferenceBlock) == 0 {
					flags |= avi.AVIIF_KEYFRAME
				}
				if err := ww.WriteData(videoStreamID, frames[i], flags); err != nil {
					return err
				}
				totalFrames++
			}
		}
	}
	var sf = avi.StreamFormat(*t.CodecPrivate)
	var handler riff.FourCC
	binary.LittleEndian.PutUint32(handler[:], sf.Compression())

	var mh avi.MainHeader
	mh.SetMicroSecPerFrame(uint32(math.Ceil(float64(*t.DefaultDuration) / 1000.0)))
	mh.SetFlags(avi.AVIF_HASINDEX | avi.AVIF_ISINTERLEAVED)
	mh.SetTotalFrames(totalFrames)
	mh.SetStreams(1) // number of audio streams plus one
	mh.SetWidth(uint32(t.Video.PixelWidth))
	mh.SetHeight(uint32(t.Video.PixelHeight))

	var sh avi.StreamHeader
	sh.SetType(avi.StreamTypeVIDS)
	sh.SetHandler(handler)
	sh.SetScale(uint32(scale))
	sh.SetRate(uint32(float64(scale) / float64(*t.DefaultDuration) * 1000000000.0))
	sh.SetLength(totalFrames)
	sh.SetSuggestedBufferSize(ww.MaxLen())

	if err := ww.WriteHeader(sh, sf); err != nil {
		return fmt.Errorf("matroska: could not write header: %w", err)
	}
	return nil
}
