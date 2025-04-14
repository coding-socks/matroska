package list

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/coding-socks/ebml"
	"github.com/coding-socks/ebml/ebmltext"
	"github.com/coding-socks/matroska"
	"github.com/coding-socks/matroska/cmd/mkc/internal/cli"
	flag "github.com/spf13/pflag"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

var Cmd = &cli.Command{
	Flags: flag.NewFlagSet("list", flag.ContinueOnError),
}

func init() {
	Cmd.Run = run
}

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
			log.Fatal(err)
		}
	}

	f, err := os.Open(args.Input)
	if err != nil {
		log.Fatalf("Could not open input file: %s", args.Input)
	}
	r := ebmltext.NewDecoder(f)
	stack := dataSizeStack{}
	printer := indentedPrinter{w: os.Stdout}

	def, _ := ebml.Definition(matroska.DocType)
	timestampScale := time.Duration(1000000) * time.Nanosecond
	trackTimestampScale := 1.0
	clusterTimestamp := time.Duration(0)
	blockDuration := time.Duration(0)

	buf := make([]byte, 8192)

	pos := 0
	for {
		id, err := r.ReadElementID()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		sel, ok := def.Get(id)
		if !ok {
			if _, err := r.Read(buf[:1]); err != nil {
				log.Fatal(err)
			}
			pos++
			continue
		}
		n := r.Release()

		ds, err := r.ReadElementDataSize()
		if err != nil {
			log.Fatal(err)
		}
		n += r.Release()
		pos += n

		if !stack.Empty() {
			p := stack.Path()
			if stack.Peek() == -1 && (!strings.HasPrefix(sel.Path, p) || sel.Path == p) {
				stack.Pop()
				printer.Sub()
			}
		}

		if sel.Type != ebml.TypeMaster {
			if int64(len(buf)) < ds {
				tmp := make([]byte, ds)
				copy(tmp, buf)
				buf = tmp
			}
		}

		printSuffix := ""
		if true {
			printSuffix += fmt.Sprintf(" at %d", pos-n)
		}
		if true {
			if ds == -1 {
				printSuffix += fmt.Sprint(", size unknown")
			} else {
				printSuffix += fmt.Sprintf(", size %d", int64(n)+ds)
			}
		}
		if true {
			if ds == -1 {
				printSuffix += fmt.Sprint(", data size unknown")
			} else {
				printSuffix += fmt.Sprintf(", data size %d", ds)
			}
		}

		skip := false
		switch id {
		default:
			switch sel.Type {
			default:
				skip = true
				printer.Print("unexpected element ", id)
			case ebml.TypeMaster:
				printer.Printf("%s%s", sel.Name, printSuffix)
			case ebml.TypeBinary:
				skip = true
				printer.Printf("%s%s", sel.Name, printSuffix)
			case ebml.TypeString:
				if _, err := io.ReadFull(r, buf[:ds]); err != nil {
					return
				}
				s, _ := ebmltext.String(buf[:ds])
				printer.Printf("%s: %s%s", sel.Name, s, printSuffix)
			case ebml.TypeUTF8:
				if _, err := io.ReadFull(r, buf[:ds]); err != nil {
					return
				}
				s, _ := ebmltext.String(buf[:ds])
				printer.Printf("%s: %s%s", sel.Name, s, printSuffix)
			case ebml.TypeUinteger:
				if _, err := io.ReadFull(r, buf[:ds]); err != nil {
					return
				}
				u, _ := ebmltext.Uint(buf[:ds])
				printer.Printf("%s: %d%s", sel.Name, u, printSuffix)
			case ebml.TypeInteger:
				if _, err := io.ReadFull(r, buf[:ds]); err != nil {
					return
				}
				i, _ := ebmltext.Int(buf[:ds])
				printer.Printf("%s: %d%s", sel.Name, i, printSuffix)
			case ebml.TypeFloat:
				if _, err := io.ReadFull(r, buf[:ds]); err != nil {
					return
				}
				f, _ := ebmltext.Float(buf[:ds])
				printer.Printf("%s: %f%s", sel.Name, f, printSuffix)
			case ebml.TypeDate:
				if _, err := io.ReadFull(r, buf[:ds]); err != nil {
					return
				}
				d, _ := ebmltext.Date(buf[:ds])
				printer.Printf("%s: %s%s", sel.Name, d, printSuffix)
			}

		case ebml.IDEBMLMaxIDLength:
			if _, err := io.ReadFull(r, buf[:ds]); err != nil {
				return
			}
			u, _ := ebmltext.Uint(buf[:ds])
			r.MaxIDLength = uint(u)
			printer.Printf("%s: %d%s", sel.Name, u, printSuffix)
		case ebml.IDEBMLMaxSizeLength:
			if _, err := io.ReadFull(r, buf[:ds]); err != nil {
				return
			}
			u, _ := ebmltext.Uint(buf[:ds])
			r.MaxSizeLength = uint(u)
			printer.Printf("%s: %d%s", sel.Name, u, printSuffix)
		case ebml.IDDocType:
			if _, err := io.ReadFull(r, buf[:ds]); err != nil {
				return
			}
			s, _ := ebmltext.String(buf[:ds])
			if s != matroska.DocType {
				log.Fatalf("unexpected document type %q, expected %q", s, matroska.DocType)
			}
			printer.Printf("%s: %s%s", sel.Name, s, printSuffix)
		case matroska.IDTimestampScale:
			if _, err := io.ReadFull(r, buf[:ds]); err != nil {
				return
			}
			u, _ := ebmltext.Uint(buf[:ds])
			timestampScale = time.Duration(u) * time.Nanosecond
			printer.Printf("%s: %d%s", sel.Name, u, printSuffix)
		case matroska.IDTrackEntry:
			trackTimestampScale = 1.0
		case matroska.IDTrackTimestampScale:
			if _, err := io.ReadFull(r, buf[:ds]); err != nil {
				return
			}
			u, _ := ebmltext.Float(buf[:ds])
			trackTimestampScale = u
			printer.Printf("%s: %d%s", sel.Name, u, printSuffix)
		case matroska.IDDuration:
			// https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-21.html#name-segment-ticks
			if _, err := io.ReadFull(r, buf[:ds]); err != nil {
				return
			}
			// Stored as a float for some reason
			f, _ := ebmltext.Float(buf[:ds])
			d := time.Duration(int64(f * float64(timestampScale)))
			printer.Printf("%s: %v%s", sel.Name, d, printSuffix)
		case matroska.IDTimestamp, matroska.IDCueDuration:
			// https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-21.html#name-segment-ticks
			if _, err := io.ReadFull(r, buf[:ds]); err != nil {
				return
			}
			i, _ := ebmltext.Uint(buf[:ds])
			d := time.Duration(i) * timestampScale
			if id == matroska.IDTimestamp {
				clusterTimestamp = d
			}
			printer.Printf("%s: %v%s", sel.Name, d, printSuffix)
		case matroska.IDBlockGroup:
			blockDuration = 0
		case matroska.IDBlockDuration, matroska.IDReferenceBlock:
			// https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-21.html#name-track-ticks
			if _, err := io.ReadFull(r, buf[:ds]); err != nil {
				return
			}
			i, _ := ebmltext.Uint(buf[:ds])
			d := time.Duration(int64(float64(i) * float64(timestampScale) * trackTimestampScale))
			blockDuration = d
			printer.Printf("%s: %v%s", sel.Name, d, printSuffix)
		case matroska.IDBlock:
			// https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-21.html#name-block-structure
			if _, err := io.ReadFull(r, buf[:ds]); err != nil {
				return
			}
			block, _ := matroska.ReadBlock(buf[:ds], clusterTimestamp)
			frames := block.Frames()
			printer.Printf("%s: track number %d, %d frame(s), timestamp %v%s", sel.Name,
				block.TrackNumber(), len(frames), blockDuration*timestampScale, printSuffix)
			printer.Add()
			for _, f := range frames {
				printer.Printf("Frame at ? size %d", len(f))
			}
			printer.Sub()
		case matroska.IDSimpleBlock:
			// https://www.ietf.org/archive/id/draft-ietf-cellar-matroska-21.html#name-simpleblock-structure
			if _, err := io.ReadFull(r, buf[:ds]); err != nil {
				return
			}
			block, _ := matroska.ReadSimpleBlock(buf[:ds], clusterTimestamp)
			printer.Printf("%s: track number %d, ? frame(s), timestamp %v%s", sel.Name,
				block.TrackNumber(), 0*timestampScale, printSuffix)
		}

		if sel.Type != ebml.TypeMaster {
			pos += int(ds)
		}

		if !stack.Empty() {
			stack.Read(n + int(ds))
			if stack.Remaining() < 1 {
				stack.Pop()
				printer.Sub()
			}
		}
		if skip {
			r.Seek(ds, io.SeekCurrent)
		} else if sel.Type == ebml.TypeMaster {
			stack.Push(sel.Path, ds)
			printer.Add()
		}
	}
}

type dataSizeStackItem struct {
	prev *dataSizeStackItem
	ds   int64
	read int
}

type dataSizeStack struct {
	item  *dataSizeStackItem
	paths []string
	lvl   int
}

func (s *dataSizeStack) Lvl() int {
	return s.lvl
}

func (s *dataSizeStack) Push(path string, ds int64) {
	s.item = &dataSizeStackItem{
		prev: s.item,
		ds:   ds,
	}
	s.paths = append(s.paths, path)
	s.lvl++
}

func (s *dataSizeStack) Pop() (path string, ds int64) {
	if s.item == nil {
		return "", -2
	}
	ds = s.item.ds
	s.item = s.item.prev
	path = s.paths[len(s.paths)-1]
	s.paths = s.paths[:len(s.paths)-1]
	s.lvl--
	return path, ds
}

func (s *dataSizeStack) Peek() int64 {
	if s.item == nil {
		return -2
	}
	return s.item.ds
}

func (s *dataSizeStack) Read(n int) {
	if s.item.ds != -1 {
		s.item.read += n
	}
}

func (s *dataSizeStack) Remaining() int {
	if s.item.ds == -1 {
		return 1
	}
	return int(s.item.ds - int64(s.item.read))
}

func (s *dataSizeStack) Empty() bool {
	return s.item == nil
}

func (s *dataSizeStack) Path() string {
	if len(s.paths) == 0 {
		return "\\"
	}
	return s.paths[len(s.paths)-1]
}

type indentedPrinter struct {
	indent []byte
	w      io.Writer
}

func (p indentedPrinter) Print(a ...any) {
	fmt.Fprintf(p.w, "%s+ %s\n", p.indent, fmt.Sprint(a...))
}

func (p indentedPrinter) Printf(format string, v ...any) {
	fmt.Fprintf(p.w, "%s+ %s\n", p.indent, fmt.Sprintf(format, v...))
}

func (p *indentedPrinter) Add() {
	if len(p.indent) == 0 {
		p.indent = append(p.indent, '|')
		return
	}
	p.indent = append(p.indent, ' ')
}

func (p *indentedPrinter) Sub() {
	p.indent = p.indent[:len(p.indent)-1]
}
