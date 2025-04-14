package extract

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/coding-socks/matroska"
	"github.com/coding-socks/matroska/cmd/mkc/internal/cli"
	flag "github.com/spf13/pflag"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var Cmd = &cli.Command{
	Flags: flag.NewFlagSet("extract", flag.ContinueOnError),
}

func init() {
	Cmd.Run = run
}

var flagOutput = Cmd.Flags.StringP("output", "o", "", "Path to the output folder.")
var flagTracks = Cmd.Flags.UintSliceP("tracks", "t", []uint{}, "Id of track to extract")

type arguments struct {
	Input  string
	Output string
	Tracks []uint
}

func run(flags *flag.FlagSet) {
	args := arguments{
		Input:  flags.Arg(0),
		Output: *flagOutput,
		Tracks: *flagTracks,
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
	if args.Output == "" {
		err := huh.NewInput().
			Title("Output folder:").
			Prompt("?").
			Validate(cli.ValidatorDir).
			Value(&args.Output).
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
	defer f.Close()

	var (
		s         *matroska.Scanner
		actionErr error
	)
	err = spinner.New().
		Title("Loading file into memory").
		Action(func() {
			s, actionErr = matroska.NewScanner(f)
		}).
		Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return
	} else if err != nil {
		log.Fatal(err)
	}

	if err = actionErr; err != nil {
		log.Fatal(err)
	}
	tracks := s.Tracks()
	if len(args.Tracks) == 0 {
		options := make([]huh.Option[uint], len(tracks.TrackEntry))
		for i, e := range tracks.TrackEntry {
			options[i] = huh.NewOption(fmt.Sprintf("Track %02d [%s]", e.TrackNumber, e.CodecID), e.TrackNumber)
		}
		err = huh.NewMultiSelect[uint]().
			Title("Track:").
			Options(options...).
			Value(&args.Tracks).
			Run()
		if errors.Is(err, huh.ErrUserAborted) {
			return
		} else if err != nil {
			log.Fatal(err)
		}
	}

	for _, trackIndex := range args.Tracks {
		te := tracks.TrackEntry[trackIndex-1]

		err = spinner.New().
			Action(func() {
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
				f, err := os.Create(filepath.Join(args.Output, fname+suffix+ext))
				if err != nil {
					actionErr = fmt.Errorf("could not create ouput file: %s", err)
					return
				}
				if err := matroska.ExtractTract(f, s, te); err != nil {
					os.Remove(filepath.Join(args.Output, fname))
					log.Fatalf("Could not extract track: %s", err)
				}
			}).
			Run()
		if errors.Is(err, huh.ErrUserAborted) {
			return
		} else if err != nil {
			log.Fatal(err)
		} else if err = actionErr; err != nil {
			log.Fatal(err)
		}
	}
}

func GuessExt(codecID string) string {
	switch codecID {
	// Audio
	case matroska.AudioCodecAAC:
		return ".aac"
	case matroska.AudioCodecAC3:
		return ".ac3"
	case matroska.AudioCodecMP2:
		return ".mp2"
	case matroska.AudioCodecMP3:
		return ".mp3"
	case matroska.AudioCodecVORBIS:
		return ".ogg"
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
