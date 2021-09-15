package list

import (
	"flag"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/coding-socks/ebml"
	"github.com/coding-socks/matroska"
	"github.com/coding-socks/matroska/cmd/mkc/internal/cli"
	"io"
	"log"
	"os"
	"strconv"
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
	var questions []*survey.Question
	if args.Input == "" {
		questions = append(questions, &survey.Question{
			Name:     "input",
			Prompt:   &survey.Input{Message: "Source matroska file:"},
			Validate: survey.ComposeValidators(survey.Required, cli.ValidatorFile),
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
	r := ebml.NewReader(f)
	for {
		el, _, err := r.Next()
		if err != nil {
			log.Fatal(err)
		}
		switch el.ID {
		default:
			log.Fatalf("ebml: unexpected element %s in root", el.ID)
		case ebml.IDVoid:
			continue
		case ebml.IDEBML:
			fmt.Print("+ EBML head\n")
		}
		r.Seek(el.DataSize.Size(), io.SeekCurrent)
		break
	}
	for {
		el, _, err := r.Next()
		if err != nil {
			log.Fatal(err)
		}
		switch el.ID {
		default:
			log.Fatalf("ebml: unexpected element %s in root", el.ID)
		case ebml.IDVoid:
			continue
		case matroska.IDSegment:
			var size string
			if el.DataSize.Known() {
				size = strconv.FormatInt(el.DataSize.Size(), 10)
			} else {
				size = "unknown"
			}
			fmt.Printf("+ Segment: size %s\n", size)
		}
		break
	}
}
