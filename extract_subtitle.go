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

func extractTrackSubtitle(w io.Writer, s *Scanner, t TrackEntry) error {
	switch t.CodecID {
	case SubtitleCodecTEXTASS, SubtitleCodecASS, SubtitleCodecTEXTSSA, SubtitleCodecSSA:
		return extractTrackSSA(w, s, t)
	case SubtitleCodecTEXTUTF8, SubtitleCodecTEXTASCII:
		return extractTrackSRT(w, s, t)
	}
	return fmt.Errorf("matroska: unknown subtitle codec %s", t.CodecID)
}

func extractTrackSRT(w io.Writer, s *Scanner, t TrackEntry) error {
	scale := s.Info().TimestampScale
	var i int
	for s.Next() {
		c := s.Cluster()
		// Due to timing and duration, SRT only uses Block Groups
		// > [...] Part 2 is used to set the timestamp of the Block, and BlockDuration element. [...]
		m := len(c.BlockGroup)
		if m == 0 {
			continue
		}
		for _, b := range c.BlockGroup {
			block, err := ReadBlock(b.Block, c.Timestamp)
			if err != nil {
				return fmt.Errorf("matroska: could not create block struct: %w", err)
			}
			if block.TrackNumber() != t.TrackNumber {
				continue
			}
			start := block.Timestamp(scale)
			var duration time.Duration
			if b.BlockDuration != nil {
				duration = time.Duration(*b.BlockDuration)
			} else if t.DefaultDuration != nil {
				duration = time.Duration(*t.DefaultDuration)
			} else {
				// TODO: https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-23.html#name-blockduration-element
				//   If a value is not present and no DefaultDuration is defined, the value is assumed to be the difference between the timestamp of this Block and the timestamp of the next Block in "display" order (not coding order).
			}
			end := block.Timestamp(scale) + duration*scale

			var sb strings.Builder

			sb.WriteString(strconv.Itoa(i + 1))
			i++
			sb.WriteRune('\n')
			sb.WriteString(subRipTime(start))
			sb.WriteString(" --> ")
			sb.WriteString(subRipTime(end))
			sb.WriteRune('\n')

			if _, err := io.WriteString(w, sb.String()); err != nil {
				return err
			}
			if _, err := io.Copy(w, block.Data()); err != nil {
				return err
			}
			if _, err := w.Write([]byte("\n\n")); err != nil {
				return err
			}
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

func extractTrackSSA(w io.Writer, s *Scanner, t TrackEntry) error {
	if _, err := w.Write(*t.CodecPrivate); err != nil {
		return err
	}

	scale := s.Info().TimestampScale
	f := subStationAlphaFormat(*t.CodecPrivate)
	events := make([]string, 0, (1<<9)-1)
	for s.Next() {
		c := s.Cluster()
		// Due to timing and duration, SSA only uses Block Groups
		// > Start & End field are used to set TimeStamp and the BlockDuration element.
		m := len(c.BlockGroup)
		if m == 0 {
			continue
		}
		for _, b := range c.BlockGroup {
			block, err := ReadBlock(b.Block, c.Timestamp)
			if err != nil {
				return fmt.Errorf("matroska: could not create block struct: %w", err)
			}
			if block.TrackNumber() != t.TrackNumber {
				continue
			}
			start := block.Timestamp(scale)
			var duration time.Duration
			if b.BlockDuration != nil {
				duration = time.Duration(*b.BlockDuration)
			} else if t.DefaultDuration != nil {
				duration = time.Duration(*t.DefaultDuration)
			} else {
				// TODO: https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-23.html#name-blockduration-element
				//   If a value is not present and no DefaultDuration is defined, the value is assumed to be the difference between the timestamp of this Block and the timestamp of the next Block in "display" order (not coding order).
			}
			end := block.Timestamp(scale) + duration*scale

			var sb strings.Builder

			sb.WriteString("Dialogue: ")

			b, _ := io.ReadAll(block.Data())
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
			events = grow(i, events)
			events[i] = sb.String()
		}
	}
	for i := range events {
		if _, err := w.Write(append([]byte(events[i]), '\n')); err != nil {
			return err
		}
	}
	return nil
}

func grow(i int, events []string) []string {
	if l := i + 1; l > len(events) {
		if l > cap(events) {
			n := cap(events)
			for l > cap(events) {
				n = (n << 1) + 1
			}
			tmp := make([]string, n)
			copy(tmp, events)
			events = tmp
		}
		events = events[:l]
	}
	return events
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
