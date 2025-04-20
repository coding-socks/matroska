package matroska

import (
	"fmt"
	"io"
)

func extractTrackVideo(w io.Writer, s *Scanner, t TrackEntry) error {
	switch t.CodecID {
	case VideoCodecMSCOMP:
		return extractTrackMSCOMP(w, s, t)
	}
	return fmt.Errorf("matroska: unknown audio codec %s", t.CodecID)
}

func extractTrackMSCOMP(w io.Writer, s *Scanner, t TrackEntry) error {
	scale := s.Info().TimestampScale
	var blocks []block
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
			blocks = append(blocks, block)
		}
		for i := range c.BlockGroup {
			block, err := ReadBlock(c.BlockGroup[i].Block, c.Timestamp)
			if err != nil {
				return fmt.Errorf("matroska: could not create block struct: %w", err)
			}
			if block.TrackNumber() != t.TrackNumber {
				continue
			}
			blocks = append(blocks, block)
		}
	}
	_ = scale
	for _, block := range blocks {
		_ = block
	}
	return nil
}
