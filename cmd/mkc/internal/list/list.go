package list

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/coding-socks/ebml"
	"github.com/coding-socks/ebml/schema"
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
	defer f.Close()

	r := io.MultiReader(f) // remove seeking capability
	s := matroska.NewScanner(r)

	v := listCallbacker{
		w: os.Stdout,
		s: s,

		printer: indentedPrinter{w: os.Stdout},

		showPosition: *flagPosition || *flagVerbose,
		showSize:     *flagSize || *flagVerbose,
		showDataSize: *flagDataSize || *flagVerbose,
	}
	s.Decoder().SetCallback(v)

	if err := s.Init(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	for s.Next() {
		s.Cluster()
	}
}

type queueCallbacker struct {
	el ebml.Element
	c  listCallbacker

	queue []func(v listCallbacker, ctx context.Context) listCallbacker
}

func (c queueCallbacker) Found(el ebml.Element, offset int64, headerSize int) ebml.Callbacker {
	f := func(cc listCallbacker, ctx context.Context) listCallbacker {
		return cc.found(el, offset, headerSize, ctx)
	}
	c.queue = append(c.queue, f)
	return c
}

func (c queueCallbacker) Decoded(el ebml.Element, offset int64, headerSize int, val any) ebml.Callbacker {
	if c.el.ID != el.ID {
		f := func(cc listCallbacker, ctx context.Context) listCallbacker {
			return cc.decode(el, offset, headerSize, val, ctx)
		}

		c.queue = append(c.queue, f)
		return c
	}
	ctx := context.WithValue(context.Background(), el.ID, val)

	for _, f := range c.queue {
		c.c = f(c.c, ctx)
	}
	return c.c
}

type listCallbacker struct {
	w io.Writer
	s *matroska.Scanner

	printer indentedPrinter
	suffix  bytes.Buffer

	showPosition bool
	showSize     bool
	showDataSize bool
}

func (c listCallbacker) Found(el ebml.Element, offset int64, headerSize int) ebml.Callbacker {
	startQueue := false

	switch el.ID {
	case matroska.IDBlockGroup:
		startQueue = true
	}

	if startQueue {
		return queueCallbacker{el: el, c: c}.Found(el, offset, headerSize)
	}

	return c.found(el, offset, headerSize, context.Background())
}

func (c listCallbacker) found(el ebml.Element, offset int64, headerSize int, ctx context.Context) listCallbacker {
	sch := el.Schema

	c.printer.Sub(offset)
	c.suffix.Reset()

	if c.showPosition {
		fmt.Fprintf(&c.suffix, ", at %d", offset)
	}
	if c.showSize {
		if el.DataSize == -1 {
			fmt.Fprint(&c.suffix, ", size unknown")
		} else {
			fmt.Fprintf(&c.suffix, ", size %d", int64(headerSize)+el.DataSize)
		}
	}
	if c.showDataSize {
		if el.DataSize == -1 {
			fmt.Fprint(&c.suffix, ", data size unknown")
		} else {
			fmt.Fprintf(&c.suffix, ", data size %d", el.DataSize)
		}
	}

	switch el.ID {
	default:
		if sch.Type != ebml.TypeMaster {
			return c
		}
		c.printer.Printf("%s%s", sch.Name, c.suffix.String())
	}

	if el.DataSize > -1 {
		absoluteEnd := offset + int64(headerSize) + el.DataSize
		c.printer.Add(absoluteEnd)
	} else {
		c.printer.Add(-1)
	}

	return c
}

func (c listCallbacker) Decoded(el ebml.Element, offset int64, headerSize int, val any) ebml.Callbacker {
	return c.decode(el, offset, headerSize, val, context.Background())
}

func (c listCallbacker) decode(el ebml.Element, offset int64, headerSize int, val any, ctx context.Context) listCallbacker {
	sch := el.Schema
	if sch.Type == ebml.TypeMaster {
		return c
	}

	c.printer.Sub(offset)
	c.suffix.Reset()

	if c.showPosition {
		fmt.Fprintf(&c.suffix, ", at %d", offset)
	}
	if c.showSize {
		if el.DataSize == -1 {
			fmt.Fprint(&c.suffix, ", size unknown")
		} else {
			fmt.Fprintf(&c.suffix, ", size %d", int64(headerSize)+el.DataSize)
		}
	}
	if c.showDataSize {
		if el.DataSize == -1 {
			fmt.Fprint(&c.suffix, ", data size unknown")
		} else {
			fmt.Fprintf(&c.suffix, ", data size %d", el.DataSize)
		}
	}

	switch el.ID {
	default:
		switch sch.Type {
		default:
			c.printer.Print("unexpected element ", el.ID)
		case ebml.TypeBinary:
			c.printer.Printf("%s%s", sch.Name, c.suffix.String())
		case ebml.TypeString:
			c.printer.Printf("%s: %s%s", sch.Name, val, c.suffix.String())
		case ebml.TypeUTF8:
			c.printer.Printf("%s: %s%s", sch.Name, val, c.suffix.String())
		case ebml.TypeUinteger:
			c.printer.Printf("%s: %d%s", sch.Name, val, c.suffix.String())
		case ebml.TypeInteger:
			c.printer.Printf("%s: %d%s", sch.Name, val, c.suffix.String())
		case ebml.TypeFloat:
			c.printer.Printf("%s: %f%s", sch.Name, val, c.suffix.String())
		case ebml.TypeDate:
			c.printer.Printf("%s: %s%s", sch.Name, val, c.suffix.String())
		}

	case matroska.IDDuration:
		// https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-21.html#name-segment-ticks
		timestampScale := c.s.Info().TimestampScale
		// Stored as a float for some reason
		f := val.(float64)
		d := time.Duration(int64(f * float64(timestampScale)))
		c.printer.Printf("%s: %v%s", sch.Name, d, c.suffix.String())
	case matroska.IDTimestamp, matroska.IDCueDuration:
		// https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-21.html#name-segment-ticks
		timestampScale := c.s.Info().TimestampScale
		i := val.(time.Duration)
		d := i * timestampScale
		c.printer.Printf("%s: %v%s", sch.Name, d, c.suffix.String())
	case matroska.IDSeekID:
		def, _ := ebml.Definition(matroska.DocType)
		i := val.(schema.ElementID)
		seekSch, _ := def.Get(i)
		c.printer.Printf("%s: %v %s%s", sch.Name, i, seekSch.Name, c.suffix.String())
	case matroska.IDBlockDuration, matroska.IDReferenceBlock:
		// https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-21.html#name-track-ticks
		timestampScale := c.s.Info().TimestampScale
		blockGroup := ctx.Value(matroska.IDBlockGroup).(matroska.BlockGroup)
		block, _ := matroska.ReadBlock(blockGroup.Block, timestampScale)

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
		for _, te := range c.s.Tracks().TrackEntry {
			if te.TrackNumber == block.TrackNumber() {
				trackTimestampScale = te.TrackTimestampScale
				break
			}
		}
		d := time.Duration(float64(i) * float64(timestampScale) * trackTimestampScale)
		c.printer.Printf("%s: %s%s", sch.Name, d, c.suffix.String())
	case matroska.IDBlock:
		// https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-21.html#name-block-structure
		timestampScale := c.s.Info().TimestampScale
		cluster := c.s.Cluster()
		codecDelay := time.Duration(0)

		b := val.([]byte)
		block, _ := matroska.ReadBlock(b, cluster.Timestamp)
		frames := block.Frames()

		for _, te := range c.s.Tracks().TrackEntry {
			if te.TrackNumber == block.TrackNumber() {
				timestampScale = time.Duration(te.TrackTimestampScale * float64(timestampScale))
				codecDelay = time.Duration(te.CodecDelay)
				break
			}
		}

		c.printer.Printf("%s: track number %d, %d frame(s), timestamp %v%s", sch.Name,
			block.TrackNumber(), len(frames), block.Timestamp(timestampScale)-codecDelay, c.suffix.String())
		c.printer.Add(1)
		frameOffset := offset + int64(headerSize) + el.DataSize
		for _, f := range frames {
			frameOffset -= int64(len(f))
		}
		for _, f := range frames {
			c.printer.Printf("Frame at %d size %d", frameOffset, len(f)) // TODO: incorrect at the moment
			frameOffset += int64(len(f))
		}
		c.printer.Sub(1)
	case matroska.IDSimpleBlock:
		// https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-21.html#name-simpleblock-structure
		timestampScale := c.s.Info().TimestampScale
		b := val.([]byte)
		block, _ := matroska.ReadSimpleBlock(b, timestampScale)
		frames := block.Frames()
		c.printer.Printf("%s: track number %d, %d frame(s), timestamp %v%s", sch.Name,
			block.TrackNumber(), len(frames), 0*timestampScale, c.suffix.String())
		c.printer.Add(1)
		frameOffset := offset + int64(headerSize) + el.DataSize
		for _, f := range frames {
			frameOffset -= int64(len(f))
		}
		for _, f := range frames {
			c.printer.Printf("Frame at %d size %d", frameOffset, len(f)) // TODO: incorrect at the moment
			frameOffset += int64(len(f))
		}
		c.printer.Sub(1)
	}

	return c
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
