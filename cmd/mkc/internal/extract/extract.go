package extract

import (
	"flag"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/briandowns/spinner"
	"github.com/coding-socks/ebml"
	"github.com/coding-socks/matroska"
	"github.com/coding-socks/matroska/cmd/mkc/internal/cli"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var Cmd = &cli.Command{
	Flags: flag.NewFlagSet("extract", flag.ContinueOnError),
}

func init() {
	Cmd.Run = run
}

var flagOutput = Cmd.Flags.String("output", "", "Path to the output folder.")

type arguments struct {
	Input  string
	Output string
}

func run(flags *flag.FlagSet) {
	args := arguments{
		Input:  flags.Arg(0),
		Output: *flagOutput,
	}
	var questions []*survey.Question
	if args.Input == "" {
		questions = append(questions, &survey.Question{
			Name:     "input",
			Prompt:   &survey.Input{Message: "Source matroska file:"},
			Validate: survey.ComposeValidators(survey.Required, cli.ValidatorFile),
		})
	}
	if args.Output == "" {
		questions = append(questions, &survey.Question{
			Name:     "output",
			Prompt:   &survey.Input{Message: "Output folder:"},
			Validate: survey.ComposeValidators(survey.Required, cli.ValidatorDir),
		})
	}
	if err := survey.Ask(questions, &args); err == terminal.InterruptErr {
		return
	} else if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(args.Input)
	if err != nil {
		log.Fatalf("Could not open input file: %s", args.Input)
	}
	defer f.Close()
	d := ebml.NewDecoder(f)
	h, err := d.DecodeHeader()
	if err != nil {
		log.Fatalf("Could not decode header: %s", err)
	}
	if h.DocType != matroska.DocType {
		log.Fatalf("File has to contain a matroska document")
	}
	spin := spinner.New(cli.SpinnerCharSet, 200*time.Millisecond)
	spin.Prefix = "Loading file into memory "
	spin.Start()
	defer spin.Stop()
	def, err := ebml.Definition(h.DocType)
	if err != nil {
		panic("Matroska document type is not registered")
	}
	var segmentEl ebml.Element
segment:
	for {
		el, _, err := d.Next()
		if err != nil {
			log.Fatal(err)
		}
		switch el.ID {
		default:
			log.Fatalf("ebml: got %s instead of segment", el.ID)
		case ebml.IDVoid:
			continue
		case def.Root.ID:
			segmentEl = el
			break segment // Done here
		}
	}
	segmentStart, _ := d.Seek(0, io.SeekCurrent)
	var offset int64
	var (
		seekHead *matroska.SeekHead
		info     *matroska.Info
		tracks   *matroska.Tracks
	)
	for {
		if end, _ := d.EndOfElement(segmentEl, offset); end {
			log.Fatal(io.ErrUnexpectedEOF)
		}
		el, n, err := d.Next()
		if segmentEl.DataSize.Known() {
			offset += int64(n)
		}
		if err == ebml.ErrInvalidVINTLength {
			continue
		} else if err == io.EOF {
			log.Fatal(io.ErrUnexpectedEOF)
		} else if err != nil {
			log.Fatal(err)
		}
		if segmentEl.DataSize.Known() {
			offset += el.DataSize.Size()
		}
		switch el.ID {
		default:
			if _, err := d.Seek(el.DataSize.Size(), io.SeekCurrent); err != nil {
				log.Fatalf("Could not skip %s: %s", el.ID, err)
			}
		case matroska.IDSeekHead:
			seekHead = &matroska.SeekHead{}
			if err := d.Decode(seekHead); err != nil {
				log.Fatalf("Could not decode %s: %s", el.ID, err)
			}
		case matroska.IDInfo:
			info = &matroska.Info{}
			if err := d.Decode(info); err != nil {
				log.Fatalf("Could not decode %s: %s", el.ID, err)
			}
		case matroska.IDTracks:
			tracks = &matroska.Tracks{}
			if err := d.Decode(tracks); err != nil {
				log.Fatalf("Could not decode %s: %s", el.ID, err)
			}
		case matroska.IDCluster:
			log.Fatalf("The first Info Element and the first Tracks Element MUST either be stored before the first Cluster Element or both SHALL be referenced by a SeekHead Element occurring before the first Cluster Element")
		}
		if seekHead != nil && info == nil {
			for _, s := range seekHead.Seek {
				id := fmt.Sprintf("0x%X", s.SeekID)
				if id == matroska.IDInfo {
					if segmentEl.DataSize.Known() {
						offset = int64(s.SeekPosition)
					}
					d.Seek(segmentStart+int64(s.SeekPosition), io.SeekStart)
					break
				}
			}
			continue
		}
		if seekHead != nil && tracks == nil {
			for _, s := range seekHead.Seek {
				id := fmt.Sprintf("0x%X", s.SeekID)
				if id == matroska.IDTracks {
					if segmentEl.DataSize.Known() {
						offset = int64(s.SeekPosition)
					}
					d.Seek(segmentStart+int64(s.SeekPosition), io.SeekStart)
					break
				}
			}
			continue
		}
		if info != nil && tracks != nil {
			break
		}
	}
	if seekHead != nil {
		for _, s := range seekHead.Seek {
			id := fmt.Sprintf("0x%X", s.SeekID)
			if id == matroska.IDCluster {
				if segmentEl.DataSize.Known() {
					offset = int64(s.SeekPosition)
				}
				d.Seek(segmentStart+int64(s.SeekPosition), io.SeekStart)
				break
			}
		}
	}
	spin.Stop()
	cs := matroska.NewClusterScanner(d, segmentEl, offset)
	options := make([]string, len(tracks.TrackEntry))
	for i, e := range tracks.TrackEntry {
		options[i] = fmt.Sprintf("Track %02d [%s]", e.TrackNumber, e.CodecID)
	}
	var trackIndex int
	err = survey.AskOne(&survey.Select{
		Message: "Select track",
		Options: options,
	}, &trackIndex)
	if err == terminal.InterruptErr {
		return
	} else if err != nil {
		log.Fatal(err)
	}
	te := tracks.TrackEntry[trackIndex]

	spin.Start()
	fname := filepath.Base(args.Input)
	fname = strings.TrimSuffix(fname, filepath.Ext(fname))
	fname = fmt.Sprintf("%s_Track_%02d", fname, te.TrackNumber)
	suffix := ""
	ext := GuessExt(te.CodecID)
	for i := 1; ; i++ {
		_, err := os.Stat(filepath.Join(args.Output, fname+suffix+ext))
		if os.IsNotExist(err) {
			break
		}
		suffix = fmt.Sprintf("_%d", i)
	}
	f, err = os.Create(filepath.Join(args.Output, fname+suffix+ext))
	if err != nil {
		log.Fatalf("Could not create ouput file: %s", err)
	}
	if err := matroska.ExtractTract(f, *info, cs, te); err != nil {
		os.Remove(filepath.Join(args.Output, fname))
		log.Fatalf("Could not extract track: %s", err)
	}
}

func GuessExt(codecID string) string {
	switch codecID {
	// Audio
	case matroska.AudioCodecAAC:
		return ".aac"
	case matroska.AudioCodecAC3:
		return ".ac3"
	case matroska.AudioCodecMP3:
		return ".mp3"
	// Video
	case matroska.VideoCodecMSCOMP:
		return ".avi"
	// Subtitle
	case matroska.SubtitleCodecTEXTASS:
		return ".ass"
	case matroska.SubtitleCodecTEXTSSA:
		return ".ssa"
	case matroska.SubtitleCodecTEXTUTF8, matroska.SubtitleCodecTEXTASCII:
		return ".srt"
	case matroska.SubtitleCodecVOBSUB, matroska.SubtitleCodecVOBSUBZLIB:
		return ".idx"
	case matroska.SubtitleCodecTEXTWEBVTT:
		return ".vtt"
	default:
		return ""
	}
}
