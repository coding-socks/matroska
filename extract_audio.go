package matroska

import (
	"fmt"
	"io"
)

func extractTrackAudio(w io.Writer, s *Scanner, t TrackEntry) error {
	switch t.CodecID {
	case AudioCodecMP2, AudioCodecMP3:
		return extractTrackMPEG(w, s, t)
	}
	return fmt.Errorf("matroska: unknown audio codec %s", t.CodecID)
}

func extractTrackMPEG(w io.Writer, s *Scanner, t TrackEntry) error {
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
			if block.TrackNumber() == t.TrackNumber {
				blocks = append(blocks, block)
			}
		}
		for i := range c.BlockGroup {
			block, err := ReadBlock(c.BlockGroup[i].Block, c.Timestamp)
			if err != nil {
				return fmt.Errorf("matroska: could not create block struct: %w", err)
			}
			if block.TrackNumber() == t.TrackNumber {
				blocks = append(blocks, block)
			}
		}
	}
	for _, block := range blocks {
		if _, err := io.Copy(w, block.Data()); err != nil {
			return err
		}
	}
	return nil
}
