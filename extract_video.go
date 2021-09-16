package matroska

import (
	"fmt"
	"io"
	"time"
)

func extractTrackVideo(w io.Writer, i Info, cs ClusterScanner, t TrackEntry) error {
	switch t.CodecID {
	case VideoCodecMSCOMP:
		return extractTrackMSCOMP(w, i, cs, t)
	}
	return fmt.Errorf("matroska: unknown audio codec %s", t.CodecID)
}

func extractTrackMSCOMP(w io.Writer, i Info, cs ClusterScanner, t TrackEntry) error {
	scale := i.TimestampScale
	var blocks []Block
	for cs.Next() {
		c := cs.Cluster()
		m := len(c.SimpleBlock) + len(c.BlockGroup)
		if m == 0 {
			continue
		}
		for i := range c.SimpleBlock {
			block, err := NewBlock(c.SimpleBlock[i], c.Timestamp, scale, 0, BlockTypeSimpleBlock)
			if err != nil {
				return fmt.Errorf("matroska: could not create block struct: %w", err)
			}
			if block.TrackNumber == int64(t.TrackNumber) {
				blocks = append(blocks, block)
			}
		}
		for i := range c.BlockGroup {
			block, err := NewBlock(c.BlockGroup[i].Block, c.Timestamp, scale, time.Duration(*c.BlockGroup[i].BlockDuration), BlockTypeBlock)
			if err != nil {
				return fmt.Errorf("matroska: could not create block struct: %w", err)
			}
			if block.TrackNumber == int64(t.TrackNumber) {
				blocks = append(blocks, block)
			}
		}
	}
	for _, block := range blocks {
		_ = block
	}
	return nil
}
