package list

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/coding-socks/ebml"
	"github.com/coding-socks/matroska"
	"github.com/coding-socks/matroska/cmd/mkc/internal/cli"
	flag "github.com/spf13/pflag"
	"io"
	"os"
	"time"
)

var Cmd = &cli.Command{
	Flags: flag.NewFlagSet("list", flag.ContinueOnError),
}

func init() {
	Cmd.Run = run
}

var flagPosition = Cmd.Flags.BoolP("position", "P", false, "Show the position of each element in decimal.")
var flagSize = Cmd.Flags.BoolP("size", "s", false, "Show the size of each element including its header.")
var flagDataSize = Cmd.Flags.BoolP("data-size", "z", false, "Show the data size of each element.")
var flagVerbose = Cmd.Flags.BoolP("verbose", "v", false, "Increase verbosity.")

type arguments struct {
	Input string
}

func run(flags *flag.FlagSet) {
	args := arguments{
		Input: flags.Arg(0),
	}
	if args.Input == "" {
		err := huh.NewInput().
			Title("Source matroska file:").
			Prompt("?").
			Validate(cli.ValidatorFile).
			Value(&args.Input).
			Run()
		if errors.Is(err, huh.ErrUserAborted) {
			return
		} else if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
	}

	f, err := os.Open(args.Input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open input file: %s\n", args.Input)
		os.Exit(1)
	}

	r := io.MultiReader(f) // remove seeking capability
	s := matroska.NewScanner(r)

	v := printVisitor{
		w: os.Stdout,
		s: s,

		printer: indentedPrinter{w: os.Stdout},

		showPosition: *flagPosition || *flagVerbose,
		showSize:     *flagSize || *flagVerbose,
		showDataSize: *flagDataSize || *flagVerbose,
	}
	s.Decoder().SetVisitor(v)

	if err := s.Init(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	for s.Next() {
		s.Cluster()
	}
}

type printVisitor struct {
	w io.Writer
	s *matroska.Scanner

	printer indentedPrinter
	suffix  bytes.Buffer

	showPosition bool
	showSize     bool
	showDataSize bool
}

func (v printVisitor) Visit(el ebml.Element, offset int64, headerSize int, val any) (w ebml.Visitor) {
	sch := el.Schema

	v.printer.Sub(offset)

	v.suffix.Reset()

	if v.showPosition {
		fmt.Fprintf(&v.suffix, ", at %d", offset)
	}
	if v.showSize {
		if el.DataSize == -1 {
			fmt.Fprint(&v.suffix, ", size unknown")
		} else {
			fmt.Fprintf(&v.suffix, ", size %d", int64(headerSize)+el.DataSize)
		}
	}
	if v.showDataSize {
		if el.DataSize == -1 {
			fmt.Fprint(&v.suffix, ", data size unknown")
		} else {
			fmt.Fprintf(&v.suffix, ", data size %d", el.DataSize)
		}
	}

	var lastTrackNumber uint = 0

	switch el.ID {
	default:
		switch sch.Type {
		default:
			v.printer.Print("unexpected element ", el.ID)
		case ebml.TypeMaster:
			v.printer.Printf("%s%s", sch.Name, v.suffix.String())
		case ebml.TypeBinary:
			v.printer.Printf("%s%s", sch.Name, v.suffix.String())
		case ebml.TypeString:
			v.printer.Printf("%s: %s%s", sch.Name, val, v.suffix.String())
		case ebml.TypeUTF8:
			v.printer.Printf("%s: %s%s", sch.Name, val, v.suffix.String())
		case ebml.TypeUinteger:
			v.printer.Printf("%s: %d%s", sch.Name, val, v.suffix.String())
		case ebml.TypeInteger:
			v.printer.Printf("%s: %d%s", sch.Name, val, v.suffix.String())
		case ebml.TypeFloat:
			v.printer.Printf("%s: %f%s", sch.Name, val, v.suffix.String())
		case ebml.TypeDate:
			v.printer.Printf("%s: %s%s", sch.Name, val, v.suffix.String())
		}
	case matroska.IDDuration:
		// https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-21.html#name-segment-ticks
		timestampScale := v.s.Info().TimestampScale
		// Stored as a float for some reason
		f := val.(float64)
		d := time.Duration(int64(f * float64(timestampScale)))
		v.printer.Printf("%s: %v%s", sch.Name, d, v.suffix.String())
	case matroska.IDTimestamp, matroska.IDCueDuration:
		// https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-21.html#name-segment-ticks
		timestampScale := v.s.Info().TimestampScale
		i := val.(time.Duration)
		d := i * timestampScale
		v.printer.Printf("%s: %v%s", sch.Name, d, v.suffix.String())
	case matroska.IDBlockDuration, matroska.IDReferenceBlock:
		// https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-21.html#name-track-ticks
		timestampScale := v.s.Info().TimestampScale
		var i int64
		switch v := val.(type) {
		default:
			panic("mkc: interface conversion failed")
		case int:
			i = int64(v)
		case uint:
			i = int64(v)
		}
		// TODO: validate if this would work. I may need the whole BlockGroup.
		trackTimestampScale := 1.0
		for _, te := range v.s.Tracks().TrackEntry {
			if te.TrackNumber == lastTrackNumber {
				trackTimestampScale = te.TrackTimestampScale
				break
			}
		}
		d := time.Duration(float64(i) * float64(timestampScale) * trackTimestampScale)
		v.printer.Printf("%s: %s%s", sch.Name, d, v.suffix.String())
	case matroska.IDBlock:
		// https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-21.html#name-block-structure
		timestampScale := v.s.Info().TimestampScale
		b := val.([]byte)
		block, _ := matroska.ReadBlock(b, timestampScale)
		frames := block.Frames()
		blockDuration := time.Second // TODO: get proper value
		v.printer.Printf("%s: track number %d, %d frame(s), timestamp %v%s", sch.Name,
			block.TrackNumber(), len(frames), blockDuration*timestampScale, v.suffix.String())
		v.printer.Add(1)
		frameOffset := offset + int64(headerSize) + el.DataSize
		for _, f := range frames {
			frameOffset -= int64(len(f))
		}
		for _, f := range frames {
			v.printer.Printf("Frame at %d size %d", frameOffset, len(f)) // TODO: incorrect at the moment
			frameOffset += int64(len(f))
		}
		v.printer.Sub(1)

		lastTrackNumber = block.TrackNumber()
	case matroska.IDSimpleBlock:
		// https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-21.html#name-simpleblock-structure
		timestampScale := v.s.Info().TimestampScale
		b := val.([]byte)
		v.printer.Printf("%s%s", sch.Name, v.suffix.String())
		block, _ := matroska.ReadSimpleBlock(b, timestampScale)
		frames := block.Frames()
		v.printer.Printf("%s: track number %d, %d frame(s), timestamp %v%s", sch.Name,
			block.TrackNumber(), len(frames), 0*timestampScale, v.suffix.String())
		v.printer.Add(1)
		frameOffset := offset + int64(headerSize) + el.DataSize
		for _, f := range frames {
			frameOffset -= int64(len(f))
		}
		for _, f := range frames {
			v.printer.Printf("Frame at %d size %d", frameOffset, len(f)) // TODO: incorrect at the moment
			frameOffset += int64(len(f))
		}
		v.printer.Sub(1)
	}

	if sch.Type == ebml.TypeMaster {
		if el.DataSize > -1 {
			absoluteEnd := offset + int64(headerSize) + el.DataSize
			v.printer.Add(absoluteEnd)
		} else {
			v.printer.Add(-1)
		}
	}

	return v
}

type indentedPrinter struct {
	indent []byte
	w      io.Writer

	ends []int64
}

func (p indentedPrinter) Print(a ...any) {
	fmt.Fprintf(p.w, "%s+ %s\n", p.indent, fmt.Sprint(a...))
}

func (p indentedPrinter) Printf(format string, v ...any) {
	fmt.Fprintf(p.w, "%s+ %s\n", p.indent, fmt.Sprintf(format, v...))
}

func (p *indentedPrinter) Add(end int64) {
	p.ends = append(p.ends, end)
	if len(p.indent) == 0 {
		p.indent = append(p.indent, '|')
		return
	}
	p.indent = append(p.indent, ' ')
}

func (p *indentedPrinter) Sub(offset int64) bool {
	end := false
	if len(p.ends) > 0 {
		for i := len(p.ends) - 1; i >= 0; i-- {
			e := p.ends[i]
			if e == -1 {
				end = true
			}
			if e <= offset {
				end = true
				p.ends = p.ends[:i]
				p.indent = p.indent[:len(p.indent)-1]
			}
		}
	}
	return end
}
