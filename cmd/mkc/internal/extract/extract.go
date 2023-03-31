package extract

import (
	"flag"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/briandowns/spinner"
	"github.com/coding-socks/matroska"
	"github.com/coding-socks/matroska/cmd/mkc/internal/cli"
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
	spin := spinner.New(cli.SpinnerCharSet, 200*time.Millisecond)
	spin.Prefix = "Loading file into memory "
	spin.Start()
	s, err := matroska.NewScanner(f)
	spin.Stop()
	if err != nil {
		log.Fatal(err)
	}
	tracks := s.Tracks()
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
	if err := matroska.ExtractTract(f, s, te); err != nil {
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
