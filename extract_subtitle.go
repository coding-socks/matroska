package matroska

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

func extractSRT(w io.Writer, i Info, cs ClusterScanner, t TrackEntry) error {
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
			block, err := NewBlock(c.BlockGroup[i].Block, c.Timestamp, scale, time.Duration(c.BlockGroup[i].BlockDuration), BlockTypeBlock)
			if err != nil {
				return fmt.Errorf("matroska: could not create block struct: %w", err)
			}
			if block.TrackNumber == int64(t.TrackNumber) {
				blocks = append(blocks, block)
			}
		}
	}
	for i, block := range blocks {
		start := block.Timestamp
		end := block.Timestamp + block.Duration

		var sb strings.Builder

		sb.WriteString(strconv.Itoa(i+1))
		sb.WriteRune('\n')
		sb.WriteString(subRipTime(start))
		sb.WriteString(" --> ")
		sb.WriteString(subRipTime(end))
		sb.WriteRune('\n')

		if _, err := io.WriteString(w, sb.String()); err != nil {
			return err
		}
		if _, err := io.Copy(w, block.Data); err != nil {
			return err
		}
		if _, err := w.Write([]byte("\n\n")); err != nil {
			return err
		}
	}
	return nil
}

func subRipTime(d time.Duration) string {
	h := d / time.Hour
	m := (d % time.Hour) / time.Minute
	s := (d % time.Minute) / time.Second
	ms := (d % time.Second) / time.Millisecond
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}

func extractSSA(w io.Writer, i Info, cs ClusterScanner, t TrackEntry) error {
	if _, err := w.Write(t.CodecPrivate); err != nil {
		return err
	}

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
			block, err := NewBlock(c.BlockGroup[i].Block, c.Timestamp, scale, time.Duration(c.BlockGroup[i].BlockDuration), BlockTypeBlock)
			if err != nil {
				return fmt.Errorf("matroska: could not create block struct: %w", err)
			}
			if block.TrackNumber == int64(t.TrackNumber) {
				blocks = append(blocks, block)
			}
		}
	}
	f := subStationAlphaFormat(t.CodecPrivate)
	orderedEvents := make([]strings.Builder, len(blocks))
	for _, block := range blocks {
		start := block.Timestamp
		end := block.Timestamp + block.Duration

		var sb strings.Builder

		sb.WriteString("Dialogue: ")

		b, _ := io.ReadAll(block.Data)
		n := len(f) + 1 // data starts with line number
		for i := range f {
			switch f[i] {
			case "marked", "start", "end":
				n--
			}
		}
		fields := strings.SplitN(string(b), ",", n)
		fieldIndex := 1 // data starts with line number
		for i := range f {
			if i > 0 {
				sb.WriteRune(',')
			}
			switch f[i] {
			case "marked":
				sb.WriteString(fmt.Sprintf("Marked=%d", 0))
			case "start":
				sb.WriteString(subStationAlphaTime(start))
			case "end":
				sb.WriteString(subStationAlphaTime(end))
			default:
				sb.WriteString(fields[fieldIndex])
				fieldIndex++
			}
		}
		i, err := strconv.Atoi(fields[0])
		if err != nil {
			return fmt.Errorf("matroska: could read SubStation Alpha line number: %w", err)
		}
		orderedEvents[i] = sb
	}
	for i := range orderedEvents {
		if _, err := w.Write(append([]byte(orderedEvents[i].String()), '\n')); err != nil {
			return err
		}
	}
	return nil
}

func subStationAlphaTime(d time.Duration) string {
	h := d / time.Hour
	m := (d % time.Hour) / time.Minute
	s := (d % time.Minute) / time.Second
	hs := (d % time.Second) / time.Millisecond / 10
	return fmt.Sprintf("%d:%02d:%02d.%02d", h, m, s, hs)
}

func subStationAlphaFormat(b []byte) []string {
	eventsSection := "[Events]"
	prefix := "Format: "
	var section string
	s := bufio.NewScanner(bytes.NewReader(b))
	for s.Scan() {
		l := s.Text()
		if section == eventsSection && strings.HasPrefix(l, prefix) {
			format := strings.Split(l[len(prefix):], ",")
			for i := range format {
				format[i] = strings.ToLower(strings.TrimSpace(format[i]))
			}
			return format
		}
		if len(l) > 1 && l[0] == '[' && l[len(l)-1] == ']' {
			section = l
		}
	}
	return nil
}
